package com.kingfisher.modules.email.task;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class EmailTask {

    public static final String TYPE_SEND_EMAIL = "email:send";

    @JsonProperty("to")
    private String to;
    @JsonProperty("subject")
    private String subject;
    @JsonProperty("body")
    private String body;

    private static final ObjectMapper MAPPER = new ObjectMapper();

    public String toJson() {
        try {
            return MAPPER.writeValueAsString(this);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public static EmailTask fromJson(String json) {
        try {
            return MAPPER.readValue(json, EmailTask.class);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public static EmailTask fromBytes(byte[] payload) {
        try {
            return MAPPER.readValue(payload, EmailTask.class);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}
