package com.kingfisher;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

@EnableAsync
@SpringBootApplication
@MapperScan("com.kingfisher.modules.**.mapper")
public class KingfisherApplication {

    public static void main(String[] args) {
        loadDotEnv();
        SpringApplication.run(KingfisherApplication.class, args);
    }

    private static void loadDotEnv() {
        List<String> candidates = List.of(".env", "../.env", "java/.env");
        for (String p : candidates) {
            Path path = Path.of(p);
            if (!Files.exists(path)) {
                continue;
            }
            try {
                for (String line : Files.readAllLines(path)) {
                    line = line.trim();
                    if (line.isEmpty() || line.startsWith("#")) {
                        continue;
                    }
                    if (line.startsWith("export ")) {
                        line = line.substring(7).trim();
                    }
                    int idx = line.indexOf('=');
                    if (idx <= 0) {
                        continue;
                    }
                    String key = line.substring(0, idx).trim();
                    String val = line.substring(idx + 1).trim();
                    if (val.length() >= 2 && ((val.startsWith("\"") && val.endsWith("\""))
                            || (val.startsWith("'") && val.endsWith("'")))) {
                        val = val.substring(1, val.length() - 1);
                    }
                    if (key.isEmpty() || System.getenv(key) != null) {
                        continue;
                    }
                    if (System.getProperty(key) == null) {
                        System.setProperty(key, val);
                    }
                }
            } catch (IOException ignored) {
            }
        }
    }
}
