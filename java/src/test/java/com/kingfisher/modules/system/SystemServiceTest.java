package com.kingfisher.modules.system;

import com.kingfisher.modules.system.service.SystemService;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class SystemServiceTest {

    @Test
    void getInfo_happy_shouldReturn() {
        SystemService service = new SystemService(null);
        var info = service.getInfo();
        assertNotNull(info);
    }

    @Test
    void getInfo_bad_withoutDeps_shouldStillReturn() {
        SystemService service = new SystemService(null);
        assertDoesNotThrow(() -> service.getInfo());
    }
}
