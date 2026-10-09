#include <jni.h>
#include <llama.h>
#include <android/log.h>
#include <algorithm>
#include <atomic>
#include <chrono>
#include <cstdint>
#include <functional>
#include <mutex>
#include <stdexcept>
#include <string>
#include <vector>

namespace {
using Clock = std::chrono::steady_clock;
int64_t elapsed(const Clock::time_point & start) { return std::chrono::duration_cast<std::chrono::milliseconds>(Clock::now() - start).count(); }
struct Result { std::string text, reason, template_hash; int prompt = 0, generated = 0; int64_t prompt_ms = 0, first_ms = -1, generation_ms = 0, total_ms = 0; };
class Stop : public std::runtime_error { public: Stop(const char * reason) : std::runtime_error(reason), reason(reason) {} const char * reason; };

class Session {
public:
    Session(const char * path, int threads, int context_tokens, int batch_tokens) {
        static std::once_flag once; std::call_once(once, [] { llama_backend_init(); });
        auto mp = llama_model_default_params(); mp.n_gpu_layers = 0;
        model = llama_model_load_from_file(path, mp); if (!model) throw std::runtime_error("Could not load GGUF model file.");
        auto cp = llama_context_default_params(); cp.n_ctx = context_tokens; cp.n_batch = batch_tokens; cp.no_perf = true;
        context = llama_init_from_model(model, cp);
        if (!context) { llama_model_free(model); model = nullptr; throw std::runtime_error("Could not create llama.cpp context."); }
        llama_set_n_threads(context, threads, threads); context_size = llama_n_ctx(context); batch_size = llama_n_batch(context);
    }
    ~Session() { if (context) llama_free(context); if (model) llama_model_free(model); }
    void cancel() { cancelled.store(true); }

    Result generate(const std::string & request_id, const std::string & system, const std::string & user,
                    int max_tokens, int64_t deadline, float temp, int top_k, float repeat, uint32_t seed, bool greedy) {
        std::lock_guard<std::mutex> lock(mutex); cancelled.store(false); const auto started = Clock::now(); Result out;
        const char * tmpl = llama_model_chat_template(model, nullptr);
        if (!tmpl || !*tmpl) return finish(out, "unsupported_template", started);
        out.template_hash = std::to_string(std::hash<std::string>{}(tmpl));
        const llama_chat_message messages[] = {{"system", system.c_str()}, {"user", user.c_str()}};
        int32_t size = llama_chat_apply_template(tmpl, messages, 2, true, nullptr, 0);
        if (size <= 0) return finish(out, "unsupported_template", started);
        std::vector<char> buffer(static_cast<size_t>(size) + 1);
        size = llama_chat_apply_template(tmpl, messages, 2, true, buffer.data(), static_cast<int32_t>(buffer.size()));
        if (size <= 0) return finish(out, "unsupported_template", started);
        std::string prompt(buffer.data(), static_cast<size_t>(size)); const llama_vocab * vocab = llama_model_get_vocab(model);
        int count = llama_tokenize(vocab, prompt.c_str(), prompt.size(), nullptr, 0, true, true); if (count < 0) count = -count;
        if (count <= 0) return finish(out, "prompt_decode_error", started);
        std::vector<llama_token> tokens(count); count = llama_tokenize(vocab, prompt.c_str(), prompt.size(), tokens.data(), count, true, true);
        if (count <= 0) return finish(out, "prompt_decode_error", started); tokens.resize(count); out.prompt = count;
        if (count + max_tokens + 8 > context_size) return finish(out, "context_overflow", started);
        llama_memory_clear(llama_get_memory(context), true); int32_t position = 0; const auto prompt_started = Clock::now();
        try { decode(tokens, position, started, deadline); }
        catch (const Stop & stop) { out.prompt_ms = elapsed(prompt_started); return finish(out, stop.reason, started); }
        catch (...) { out.prompt_ms = elapsed(prompt_started); return finish(out, "prompt_decode_error", started); }
        out.prompt_ms = elapsed(prompt_started);
        llama_sampler * sampler = llama_sampler_chain_init(llama_sampler_chain_default_params());
        if (greedy) llama_sampler_chain_add(sampler, llama_sampler_init_greedy());
        else {
            llama_sampler_chain_add(sampler, llama_sampler_init_top_k(top_k));
            llama_sampler_chain_add(sampler, llama_sampler_init_penalties(llama_vocab_n_tokens(vocab), 64, repeat, 0, 0));
            llama_sampler_chain_add(sampler, llama_sampler_init_temp(temp)); llama_sampler_chain_add(sampler, llama_sampler_init_dist(seed));
        }
        const auto generation_started = Clock::now(); out.reason = "max_tokens";
        try {
            for (int i = 0; i < max_tokens; ++i) {
                check(started, deadline); llama_token token = llama_sampler_sample(sampler, context, -1); llama_sampler_accept(sampler, token);
                if (llama_vocab_is_eog(vocab, token)) { out.reason = "eog"; break; }
                if (out.first_ms < 0) out.first_ms = elapsed(started); char piece[512];
                int length = llama_token_to_piece(vocab, token, piece, sizeof(piece), 0, false); if (length > 0) out.text.append(piece, piece + length);
                decode(std::vector<llama_token>{token}, position, started, deadline); ++out.generated;
            }
        } catch (const Stop & stop) { out.reason = stop.reason; } catch (...) { out.reason = "native_error"; }
        llama_sampler_free(sampler); out.generation_ms = elapsed(generation_started); out.total_ms = elapsed(started);
        __android_log_print(ANDROID_LOG_INFO, "GmnLlamaNative", "requestId=%s stop=%s promptTokens=%d generatedTokens=%d", request_id.c_str(), out.reason.c_str(), out.prompt, out.generated);
        return out;
    }
private:
    Result finish(Result out, const char * reason, const Clock::time_point & started) { out.reason = reason; out.total_ms = elapsed(started); return out; }
    void check(const Clock::time_point & started, int64_t deadline) const { if (cancelled.load()) throw Stop("cancelled"); if (deadline > 0 && elapsed(started) >= deadline) throw Stop("timeout"); }
    void decode(const std::vector<llama_token> & tokens, int32_t & position, const Clock::time_point & started, int64_t deadline) {
        llama_batch batch = llama_batch_init(batch_size, 0, 1);
        try { for (size_t offset = 0; offset < tokens.size(); offset += batch_size) { check(started, deadline); batch.n_tokens = 0; size_t n = std::min(static_cast<size_t>(batch_size), tokens.size() - offset);
            for (size_t i = 0; i < n; ++i) { size_t index = offset + i; batch.token[batch.n_tokens] = tokens[index]; batch.pos[batch.n_tokens] = position++; batch.n_seq_id[batch.n_tokens] = 1; batch.seq_id[batch.n_tokens][0] = 0; batch.logits[batch.n_tokens] = index == tokens.size() - 1; ++batch.n_tokens; }
            if (llama_decode(context, batch) != 0) throw std::runtime_error("llama_decode failed"); }
            llama_batch_free(batch); } catch (...) { llama_batch_free(batch); throw; }
    }
    llama_model * model = nullptr; llama_context * context = nullptr; int32_t context_size = 0, batch_size = 0; std::atomic_bool cancelled = false; std::mutex mutex;
};

void illegal(JNIEnv * env, const char * message) { env->ThrowNew(env->FindClass("java/lang/IllegalStateException"), message); }
jobjectArray array(JNIEnv * env, const Result & r) {
    std::string values[] = {r.text, r.reason, std::to_string(r.prompt), std::to_string(r.generated), std::to_string(r.prompt_ms), std::to_string(r.first_ms), std::to_string(r.generation_ms), std::to_string(r.total_ms), r.template_hash};
    jclass type = env->FindClass("java/lang/String"); jobjectArray out = env->NewObjectArray(9, type, nullptr);
    for (int i = 0; i < 9; ++i) { jstring value = env->NewStringUTF(values[i].c_str()); env->SetObjectArrayElement(out, i, value); env->DeleteLocalRef(value); } return out;
}
} // namespace

extern "C" JNIEXPORT jlong JNICALL Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeLoad(JNIEnv * env, jobject, jstring path_value, jint threads, jint context, jint batch) {
    const char * path = env->GetStringUTFChars(path_value, nullptr); try { auto * session = new Session(path, threads, context, batch); env->ReleaseStringUTFChars(path_value, path); return reinterpret_cast<jlong>(session); }
    catch (const std::exception & e) { env->ReleaseStringUTFChars(path_value, path); illegal(env, e.what()); return 0; }
}
extern "C" JNIEXPORT jobjectArray JNICALL Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeGenerate(JNIEnv * env, jobject, jlong handle, jstring request_value, jstring system_value, jstring user_value, jint max_tokens, jlong deadline, jfloat temp, jint top_k, jfloat repeat, jint seed, jboolean greedy) {
    auto * session = reinterpret_cast<Session *>(handle); if (!session) { illegal(env, "No llama.cpp model is loaded."); return nullptr; }
    const char * request = env->GetStringUTFChars(request_value, nullptr); const char * system = env->GetStringUTFChars(system_value, nullptr); const char * user = env->GetStringUTFChars(user_value, nullptr);
    try { Result result = session->generate(request, system, user, max_tokens, deadline, temp, top_k, repeat, static_cast<uint32_t>(seed), greedy); env->ReleaseStringUTFChars(request_value, request); env->ReleaseStringUTFChars(system_value, system); env->ReleaseStringUTFChars(user_value, user); return array(env, result); }
    catch (const std::exception & e) { env->ReleaseStringUTFChars(request_value, request); env->ReleaseStringUTFChars(system_value, system); env->ReleaseStringUTFChars(user_value, user); illegal(env, e.what()); return nullptr; }
}
extern "C" JNIEXPORT void JNICALL Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeCancel(JNIEnv *, jobject, jlong handle) { auto * session = reinterpret_cast<Session *>(handle); if (session) session->cancel(); }
extern "C" JNIEXPORT void JNICALL Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeUnload(JNIEnv *, jobject, jlong handle) { delete reinterpret_cast<Session *>(handle); }
