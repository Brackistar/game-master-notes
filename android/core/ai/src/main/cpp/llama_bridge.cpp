#include <jni.h>
#include <llama.h>
#include <android/log.h>

#include <atomic>
#include <algorithm>
#include <cstdint>
#include <chrono>
#include <functional>
#include <mutex>
#include <stdexcept>
#include <string>
#include <vector>

namespace {

constexpr const char * LOG_TAG = "GmnLlamaNative";

#define GMN_LOGI(...) __android_log_print(ANDROID_LOG_INFO, LOG_TAG, __VA_ARGS__)
#define GMN_LOGW(...) __android_log_print(ANDROID_LOG_WARN, LOG_TAG, __VA_ARGS__)
#define GMN_LOGE(...) __android_log_print(ANDROID_LOG_ERROR, LOG_TAG, __VA_ARGS__)

int64_t elapsed_millis(const std::chrono::steady_clock::time_point & started_at) {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
        std::chrono::steady_clock::now() - started_at
    ).count();
}

class LlamaSession {
public:
    LlamaSession(const char * model_path, int thread_count, int context_tokens) {
        const auto started_at = std::chrono::steady_clock::now();
        GMN_LOGI("Session load started pathHash=%zu threads=%d requestedContext=%d", std::hash<std::string>{}(model_path), thread_count, context_tokens);
        static std::once_flag backend_once;
        std::call_once(backend_once, [] {
            GMN_LOGI("llama_backend_init started");
            llama_backend_init();
            GMN_LOGI("llama_backend_init finished");
        });

        llama_model_params model_params = llama_model_default_params();
        model_params.n_gpu_layers = 0;

        model_ = llama_model_load_from_file(model_path, model_params);
        if (model_ == nullptr) {
            GMN_LOGE("Model load failed pathHash=%zu", std::hash<std::string>{}(model_path));
            throw std::runtime_error("Could not load GGUF model file.");
        }
        GMN_LOGI("Model file loaded elapsedMs=%lld", static_cast<long long>(elapsed_millis(started_at)));

        llama_context_params context_params = llama_context_default_params();
        context_params.n_ctx = static_cast<uint32_t>(context_tokens);
        context_params.n_batch = BATCH_TOKENS;
        context_params.no_perf = true;

        context_ = llama_init_from_model(model_, context_params);
        if (context_ == nullptr) {
            llama_model_free(model_);
            model_ = nullptr;
            throw std::runtime_error("Could not create llama.cpp context.");
        }

        llama_set_n_threads(context_, thread_count, thread_count);
        context_tokens_ = llama_n_ctx(context_);
        batch_tokens_ = llama_n_batch(context_);

        llama_sampler_chain_params sampler_params = llama_sampler_chain_default_params();
        sampler_ = llama_sampler_chain_init(sampler_params);
        llama_sampler_chain_add(sampler_, llama_sampler_init_greedy());
        GMN_LOGI(
            "Session load finished actualContext=%d batchTokens=%d elapsedMs=%lld",
            context_tokens_,
            batch_tokens_,
            static_cast<long long>(elapsed_millis(started_at))
        );
    }

    ~LlamaSession() {
        GMN_LOGI("Session unload started");
        const auto started_at = std::chrono::steady_clock::now();
        if (sampler_ != nullptr) {
            llama_sampler_free(sampler_);
        }
        if (context_ != nullptr) {
            llama_free(context_);
        }
        if (model_ != nullptr) {
            llama_model_free(model_);
        }
        GMN_LOGI("Session unload finished elapsedMs=%lld", static_cast<long long>(elapsed_millis(started_at)));
    }

    std::string generate(const std::string & prompt, int max_tokens, int64_t max_duration_millis) {
        cancelled_.store(false);
        const auto started_at = std::chrono::steady_clock::now();
        GMN_LOGI(
            "Generate started promptChars=%zu maxTokens=%d timeoutMs=%lld contextTokens=%d batchTokens=%d",
            prompt.size(),
            max_tokens,
            static_cast<long long>(max_duration_millis),
            context_tokens_,
            batch_tokens_
        );

        const llama_vocab * vocab = llama_model_get_vocab(model_);
        const auto tokenize_started_at = std::chrono::steady_clock::now();
        int token_count = llama_tokenize(
            vocab,
            prompt.c_str(),
            static_cast<int32_t>(prompt.size()),
            nullptr,
            0,
            true,
            true
        );
        if (token_count < 0) {
            token_count = -token_count;
        }
        if (token_count == 0) {
            throw std::runtime_error("Prompt did not produce tokens.");
        }

        std::vector<llama_token> tokens(static_cast<size_t>(token_count));
        const int actual_token_count = llama_tokenize(
            vocab,
            prompt.c_str(),
            static_cast<int32_t>(prompt.size()),
            tokens.data(),
            token_count,
            true,
            true
        );
        if (actual_token_count < 0) {
            throw std::runtime_error("Prompt tokenization failed.");
        }
        tokens.resize(static_cast<size_t>(actual_token_count));
        const int prompt_tokens_before_trim = static_cast<int>(tokens.size());
        trim_to_context(tokens, max_tokens);
        GMN_LOGI(
            "Prompt tokenized promptTokens=%d promptTokensAfterTrim=%zu tokenizeElapsedMs=%lld",
            prompt_tokens_before_trim,
            tokens.size(),
            static_cast<long long>(elapsed_millis(tokenize_started_at))
        );

        llama_memory_clear(llama_get_memory(context_), true);

        int32_t current_position = 0;
        const auto prompt_decode_started_at = std::chrono::steady_clock::now();
        decode_tokens(tokens, current_position, true, started_at, max_duration_millis, "prompt");
        GMN_LOGI(
            "Prompt decoded promptTokens=%zu elapsedMs=%lld",
            tokens.size(),
            static_cast<long long>(elapsed_millis(prompt_decode_started_at))
        );

        std::string output;
        llama_sampler_reset(sampler_);

        int generated_tokens = 0;
        const char * stop_reason = "max_tokens";
        const auto token_generation_started_at = std::chrono::steady_clock::now();
        for (int i = 0; i < max_tokens && !cancelled_.load(); ++i) {
            if (is_timed_out(started_at, max_duration_millis)) {
                stop_reason = "timeout";
                break;
            }
            const llama_token token = llama_sampler_sample(sampler_, context_, -1);
            llama_sampler_accept(sampler_, token);
            if (llama_vocab_is_eog(vocab, token)) {
                stop_reason = "eog";
                break;
            }

            char piece[256];
            const int piece_length = llama_token_to_piece(vocab, token, piece, sizeof(piece), 0, false);
            if (piece_length > 0) {
                output.append(piece, piece + piece_length);
            }

            const std::vector<llama_token> next_tokens = { token };
            decode_tokens(next_tokens, current_position, true, started_at, max_duration_millis, "generation");
            generated_tokens++;
        }
        if (cancelled_.load()) {
            stop_reason = "cancelled";
        }

        GMN_LOGI(
            "Generate finished stopReason=%s generatedTokens=%d outputChars=%zu tokenLoopElapsedMs=%lld totalElapsedMs=%lld",
            stop_reason,
            generated_tokens,
            output.size(),
            static_cast<long long>(elapsed_millis(token_generation_started_at)),
            static_cast<long long>(elapsed_millis(started_at))
        );

        return output;
    }

    void cancel() {
        GMN_LOGW("Cancel flag set");
        cancelled_.store(true);
    }

private:
    static bool is_timed_out(
        const std::chrono::steady_clock::time_point & started_at,
        int64_t max_duration_millis
    ) {
        if (max_duration_millis <= 0) return false;
        const auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::steady_clock::now() - started_at
        ).count();
        return elapsed >= max_duration_millis;
    }

    void trim_to_context(std::vector<llama_token> & tokens, int max_tokens) const {
        const int32_t reserved_response_tokens = std::max(1, max_tokens);
        const int32_t prompt_budget = std::max(
            1,
            context_tokens_ - reserved_response_tokens - CONTEXT_HEADROOM
        );
        if (static_cast<int32_t>(tokens.size()) > prompt_budget) {
            GMN_LOGW(
                "Prompt trimmed originalTokens=%zu promptBudget=%d reservedResponseTokens=%d",
                tokens.size(),
                prompt_budget,
                reserved_response_tokens
            );
            tokens.erase(tokens.begin(), tokens.end() - prompt_budget);
        }
    }

    void decode_tokens(
        const std::vector<llama_token> & tokens,
        int32_t & current_position,
        bool compute_last_logit,
        const std::chrono::steady_clock::time_point & started_at,
        int64_t max_duration_millis,
        const char * phase
    ) {
        llama_batch batch = llama_batch_init(batch_tokens_, 0, 1);
        try {
            for (size_t offset = 0; offset < tokens.size(); offset += batch_tokens_) {
                if (cancelled_.load()) {
                    GMN_LOGW("Decode cancelled phase=%s offset=%zu totalTokens=%zu", phase, offset, tokens.size());
                    throw std::runtime_error("llama.cpp generation was cancelled.");
                }
                if (is_timed_out(started_at, max_duration_millis)) {
                    GMN_LOGW(
                        "Decode timed out phase=%s offset=%zu totalTokens=%zu timeoutMs=%lld",
                        phase,
                        offset,
                        tokens.size(),
                        static_cast<long long>(max_duration_millis)
                    );
                    throw std::runtime_error("llama.cpp prompt evaluation timed out.");
                }
                batch.n_tokens = 0;
                const size_t batch_size = std::min(
                    static_cast<size_t>(batch_tokens_),
                    tokens.size() - offset
                );

                for (size_t i = 0; i < batch_size; ++i) {
                    const size_t token_index = offset + i;
                    batch.token[batch.n_tokens] = tokens[token_index];
                    batch.pos[batch.n_tokens] = current_position++;
                    batch.n_seq_id[batch.n_tokens] = 1;
                    batch.seq_id[batch.n_tokens][0] = 0;
                    batch.logits[batch.n_tokens] =
                        compute_last_logit && token_index == tokens.size() - 1;
                    batch.n_tokens++;
                }

                const auto batch_started_at = std::chrono::steady_clock::now();
                GMN_LOGI(
                    "Decode batch started phase=%s offset=%zu batchSize=%zu currentPosition=%d computeLastLogit=%d",
                    phase,
                    offset,
                    batch_size,
                    current_position,
                    compute_last_logit ? 1 : 0
                );
                if (llama_decode(context_, batch) != 0) {
                    throw std::runtime_error("llama.cpp failed to decode tokens.");
                }
                GMN_LOGI(
                    "Decode batch finished phase=%s offset=%zu batchSize=%zu elapsedMs=%lld",
                    phase,
                    offset,
                    batch_size,
                    static_cast<long long>(elapsed_millis(batch_started_at))
                );
            }
            llama_batch_free(batch);
        } catch (...) {
            llama_batch_free(batch);
            throw;
        }
    }

    llama_model * model_ = nullptr;
    llama_context * context_ = nullptr;
    llama_sampler * sampler_ = nullptr;
    int32_t context_tokens_ = 0;
    int32_t batch_tokens_ = 0;
    std::atomic_bool cancelled_ = false;

    static constexpr int32_t BATCH_TOKENS = 32;
    static constexpr int32_t CONTEXT_HEADROOM = 8;
};

jstring to_jstring(JNIEnv * env, const std::string & value) {
    return env->NewStringUTF(value.c_str());
}

void throw_illegal_state(JNIEnv * env, const std::string & message) {
    jclass exception_class = env->FindClass("java/lang/IllegalStateException");
    env->ThrowNew(exception_class, message.c_str());
}

} // namespace

extern "C" JNIEXPORT jlong JNICALL
Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeLoad(
    JNIEnv * env,
    jobject,
    jstring model_path,
    jint thread_count,
    jint context_tokens
) {
    const char * path = env->GetStringUTFChars(model_path, nullptr);
    try {
        auto * session = new LlamaSession(path, thread_count, context_tokens);
        env->ReleaseStringUTFChars(model_path, path);
        return reinterpret_cast<jlong>(session);
    } catch (const std::exception & error) {
        GMN_LOGE("nativeLoad failed: %s", error.what());
        env->ReleaseStringUTFChars(model_path, path);
        throw_illegal_state(env, error.what());
        return 0;
    }
}

extern "C" JNIEXPORT jstring JNICALL
Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeGenerate(
    JNIEnv * env,
    jobject,
    jlong handle,
    jstring prompt,
    jint max_tokens,
    jlong max_duration_millis
) {
    auto * session = reinterpret_cast<LlamaSession *>(handle);
    if (session == nullptr) {
        throw_illegal_state(env, "No llama.cpp model is loaded.");
        return nullptr;
    }

    const char * prompt_text = env->GetStringUTFChars(prompt, nullptr);
    try {
        std::string output = session->generate(prompt_text, max_tokens, max_duration_millis);
        env->ReleaseStringUTFChars(prompt, prompt_text);
        return to_jstring(env, output);
    } catch (const std::exception & error) {
        GMN_LOGE("nativeGenerate failed: %s", error.what());
        env->ReleaseStringUTFChars(prompt, prompt_text);
        throw_illegal_state(env, error.what());
        return nullptr;
    }
}

extern "C" JNIEXPORT void JNICALL
Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeCancel(
    JNIEnv *,
    jobject,
    jlong handle
) {
    auto * session = reinterpret_cast<LlamaSession *>(handle);
    if (session != nullptr) {
        session->cancel();
    }
}

extern "C" JNIEXPORT void JNICALL
Java_com_brackistar_gamemasternotes_core_ai_LlamaCppBridge_nativeUnload(
    JNIEnv *,
    jobject,
    jlong handle
) {
    auto * session = reinterpret_cast<LlamaSession *>(handle);
    delete session;
}
