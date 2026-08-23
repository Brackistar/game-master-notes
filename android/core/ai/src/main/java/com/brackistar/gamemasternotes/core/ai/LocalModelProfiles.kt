package com.brackistar.gamemasternotes.core.ai

data class LocalModelProfile(
    val model: AiModel,
    val family: LocalModelFamily,
    val promptStyle: PromptStyle,
    val modelFileName: String,
)

enum class LocalModelFamily {
    Lfm25,
    Qwen3,
    Gemma3,
    Phi4,
}

enum class PromptStyle {
    Plain,
    ChatMl,
    Gemma,
    Phi,
}

object LocalModelProfiles {
    val Lfm25TinyQ2 = LocalModelProfile(
        model = AiModel(
            id = "lfm2.5-350m-q2-tiny",
            displayName = "LFM2.5 350M Tiny Q2",
            fileSizeBytes = 161_000_000L,
            quantization = "Q2_K",
            minimumRamMb = 2_048,
            description = "Smallest LFM2.5 slot for very slow tablets.",
        ),
        family = LocalModelFamily.Lfm25,
        promptStyle = PromptStyle.Plain,
        modelFileName = "lfm2.5-350m-tiny-q2.gguf",
    )

    val Lfm25 = LocalModelProfile(
        model = AiModel(
            id = "lfm2.5-350m-q4",
            displayName = "LFM2.5 350M Balanced",
            fileSizeBytes = 229_000_000L,
            quantization = "Q4_K_M",
            minimumRamMb = 2_048,
            description = "Tiny baseline model for low-RAM Android devices.",
        ),
        family = LocalModelFamily.Lfm25,
        promptStyle = PromptStyle.Plain,
        modelFileName = "lfm2.5-350m-q4.gguf",
    )

    val Lfm25FastQ4 = LocalModelProfile(
        model = AiModel(
            id = "lfm2.5-350m-q4-fast",
            displayName = "LFM2.5 350M Fast Q4",
            fileSizeBytes = 229_000_000L,
            quantization = "Q4_0 or Q4_K_S",
            minimumRamMb = 2_048,
            description = "Speed-oriented LFM2.5 slot for tablet testing.",
        ),
        family = LocalModelFamily.Lfm25,
        promptStyle = PromptStyle.Plain,
        modelFileName = "lfm2.5-350m-fast-q4.gguf",
    )

    val Lfm25TinyQ3 = LocalModelProfile(
        model = AiModel(
            id = "lfm2.5-350m-q3-tiny",
            displayName = "LFM2.5 350M Tiny Q3",
            fileSizeBytes = 180_000_000L,
            quantization = "Q3_K_M or Q3_K_S",
            minimumRamMb = 2_048,
            description = "Lowest-latency LFM2.5 slot when Q4 is still too slow.",
        ),
        family = LocalModelFamily.Lfm25,
        promptStyle = PromptStyle.Plain,
        modelFileName = "lfm2.5-350m-tiny-q3.gguf",
    )

    val Gemma3 = LocalModelProfile(
        model = AiModel(
            id = "gemma-3-1b-it-q4",
            displayName = "Gemma 3 1B Instruct",
            fileSizeBytes = 800_000_000L,
            quantization = "Q4",
            minimumRamMb = 3_072,
            description = "Smallest practical local model fallback for weak devices.",
        ),
        family = LocalModelFamily.Gemma3,
        promptStyle = PromptStyle.Gemma,
        modelFileName = "gemma-3-1b-it-q4.gguf",
    )

    val Qwen3 = LocalModelProfile(
        model = AiModel(
            id = "qwen3-1.7b-instruct-q4",
            displayName = "Qwen3 1.7B Instruct",
            fileSizeBytes = 1_200_000_000L,
            quantization = "Q4",
            minimumRamMb = 4_096,
            description = "Best first quality target for low-RAM RAG answers.",
        ),
        family = LocalModelFamily.Qwen3,
        promptStyle = PromptStyle.ChatMl,
        modelFileName = "qwen3-1.7b-instruct-q4.gguf",
    )

    val Phi4 = LocalModelProfile(
        model = AiModel(
            id = "phi-4-mini-instruct-q4",
            displayName = "Phi-4 Mini Instruct",
            fileSizeBytes = 2_700_000_000L,
            quantization = "Q4",
            minimumRamMb = 6_144,
            description = "Higher quality option when the device has enough memory.",
        ),
        family = LocalModelFamily.Phi4,
        promptStyle = PromptStyle.Phi,
        modelFileName = "phi-4-mini-instruct-q4.gguf",
    )

    val Lfm25Family = listOf(Lfm25TinyQ2, Lfm25TinyQ3, Lfm25FastQ4, Lfm25)
    val All = Lfm25Family + listOf(Gemma3, Qwen3, Phi4)
}
