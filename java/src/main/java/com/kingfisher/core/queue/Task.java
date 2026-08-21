package com.kingfisher.core.queue;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Task {
    @JsonProperty("type")
    private String type;
    @JsonProperty("payload")
    private String payload;

    public static Task of(String type, String payload) {
        return new Task(type, payload);
    }

    public static Task ofJson(String type, Object payloadObj) {
        try {
            com.fasterxml.jackson.databind.ObjectMapper m = new com.fasterxml.jackson.databind.ObjectMapper();
            return new Task(type, m.writeValueAsString(payloadObj));
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}
