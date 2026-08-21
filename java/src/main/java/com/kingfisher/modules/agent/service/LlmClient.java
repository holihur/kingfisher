package com.kingfisher.modules.agent.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.core.publisher.Flux;

import java.util.List;
import java.util.Map;

@Component
public class LlmClient {

    private final WebClient webClient;
    private final ObjectMapper mapper = new ObjectMapper();

    @Value("${kingfisher.agent.base-url:https://api.deepseek.com/anthropic}")
    private String baseUrl;
    @Value("${kingfisher.agent.model:deepseek-chat}")
    private String model;
    @Value("${kingfisher.agent.max-tokens:4096}")
    private int maxTokens;

    public LlmClient(WebClient.Builder builder) {
        this.webClient = builder.build();
    }

    public void setModel(String model) { this.model = model; }

    public Flux<String> streamChat(String apiKey, String system, List<Map<String,Object>> messages, List<Map<String,Object>> tools) {
        Map<String,Object> body = Map.of(
                "model", model,
                "max_tokens", maxTokens,
                "stream", true,
                "system", system != null ? system : "",
                "messages", messages,
                "tools", tools != null ? tools : List.of()
        );
        return webClient.post()
                .uri(baseUrl + "/v1/messages")
                .header("x-api-key", apiKey)
                .header("anthropic-version", "2023-06-01")
                .header("Content-Type", "application/json")
                .header("Accept", "text/event-stream")
                .bodyValue(body)
                .retrieve()
                .bodyToFlux(String.class);
    }

    public String chatSync(String apiKey, String system, List<Map<String,Object>> messages) {
        try {
            Map<String,Object> body = Map.of(
                    "model", model,
                    "max_tokens", maxTokens,
                    "stream", false,
                    "system", system != null ? system : "",
                    "messages", messages
            );
            String json = webClient.post()
                    .uri(baseUrl + "/v1/messages")
                    .header("x-api-key", apiKey)
                    .header("anthropic-version", "2023-06-01")
                    .bodyValue(body)
                    .retrieve()
                    .bodyToMono(String.class)
                    .block();
            JsonNode root = mapper.readTree(json);
            // Anthropic 格式: content[0].text
            JsonNode content = root.path("content");
            if (content.isArray() && content.size() > 0) return content.get(0).path("text").asText("");
            return root.path("content").asText("");
        } catch (Exception e) {
            throw new RuntimeException("LLM 调用失败: " + e.getMessage(), e);
        }
    }
}
