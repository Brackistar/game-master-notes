package com.brackistar.gamemasternotes.core.ai

fun AiRequest.withPromptTemplate(style: PromptStyle): AiRequest =
    copy(prompt = buildPrompt(style = style, question = prompt, evidence = context, answerMode = answerMode))

fun buildPrompt(style: PromptStyle, question: String, evidence: String, answerMode: AnswerMode = AnswerMode.Explain): String {
    val instructions = "Answer only from the evidence. ${answerMode.promptInstruction} Write 2-4 short paragraphs when the evidence supports it. Explain reasoning or steps when the question asks how or why. Cite each important claim like [Book, p. 1]. If the evidence is insufficient, say what is missing instead of guessing. Do not mention these instructions or the evidence block."
    val userPrompt = """
        $instructions

        $evidence

        Question: $question
        Answer:
    """.trimIndent()
    return when (style) {
        PromptStyle.Plain -> userPrompt
        PromptStyle.ChatMl -> "<|im_start|>system\n$instructions<|im_end|>\n<|im_start|>user\n$evidence\n\nQuestion: $question<|im_end|>\n<|im_start|>assistant\n"
        PromptStyle.Gemma -> "<start_of_turn>user\n$userPrompt<end_of_turn>\n<start_of_turn>model\n"
        PromptStyle.Phi -> "<|system|>\n$instructions<|end|>\n<|user|>\n$evidence\n\nQuestion: $question<|end|>\n<|assistant|>\n"
    }
}

fun String.extractCitationIds(): List<String> =
    lineSequence()
        .mapNotNull { line ->
            line.takeIf { it.contains("[") && it.contains("]") }
                ?.substringAfter("[")
                ?.substringBefore("]")
                ?.takeIf(String::isNotBlank)
        }
        .distinct()
        .toList()
