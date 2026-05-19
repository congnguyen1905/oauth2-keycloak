package com.example.oauth2demo.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

/**
 * Resource Server REST API
 * 
 * Các endpoint này được bảo vệ bằng JWT Bearer token.
 * Client gọi API phải gửi header: Authorization: Bearer <access_token>
 */
@RestController
@RequestMapping("/api")
public class ApiController {

    /**
     * Public endpoint - không cần auth
     */
    @GetMapping("/public/hello")
    public ResponseEntity<Map<String, String>> publicHello() {
        Map<String, String> response = new HashMap<>();
        response.put("message", "Hello from Public API - no auth required!");
        response.put("status", "public");
        return ResponseEntity.ok(response);
    }

    /**
     * Protected endpoint - cần JWT token hợp lệ
     */
    @GetMapping("/user/profile")
    public ResponseEntity<Map<String, Object>> userProfile(@AuthenticationPrincipal Jwt jwt) {
        Map<String, Object> profile = new HashMap<>();
        profile.put("username", jwt.getClaimAsString("preferred_username"));
        profile.put("email", jwt.getClaimAsString("email"));
        profile.put("firstName", jwt.getClaimAsString("given_name"));
        profile.put("lastName", jwt.getClaimAsString("family_name"));
        profile.put("subject", jwt.getSubject());
        profile.put("issuedAt", jwt.getIssuedAt());
        profile.put("expiresAt", jwt.getExpiresAt());
        profile.put("roles", jwt.getClaimAsStringList("roles"));
        return ResponseEntity.ok(profile);
    }

    /**
     * Protected endpoint - cần ROLE_USER
     */
    @GetMapping("/user/data")
    @PreAuthorize("hasRole('USER')")
    public ResponseEntity<Map<String, Object>> userData(@AuthenticationPrincipal Jwt jwt) {
        Map<String, Object> data = new HashMap<>();
        data.put("message", "Protected user data - requires ROLE_USER");
        data.put("username", jwt.getClaimAsString("preferred_username"));
        data.put("data", "Đây là dữ liệu nhạy cảm chỉ dành cho user đã xác thực.");
        return ResponseEntity.ok(data);
    }

    /**
     * Admin endpoint - chỉ ROLE_ADMIN mới truy cập được
     */
    @GetMapping("/admin/stats")
    @PreAuthorize("hasRole('ADMIN')")
    public ResponseEntity<Map<String, Object>> adminStats(@AuthenticationPrincipal Jwt jwt) {
        Map<String, Object> stats = new HashMap<>();
        stats.put("message", "Admin-only endpoint - requires ROLE_ADMIN");
        stats.put("calledBy", jwt.getClaimAsString("preferred_username"));
        stats.put("totalUsers", 2);
        stats.put("activeTokens", 1);
        stats.put("serverStatus", "healthy");
        return ResponseEntity.ok(stats);
    }

    /**
     * Endpoint để xem raw JWT claims - hữu ích khi debug
     */
    @GetMapping("/token/info")
    public ResponseEntity<Map<String, Object>> tokenInfo(@AuthenticationPrincipal Jwt jwt) {
        Map<String, Object> tokenInfo = new HashMap<>();
        tokenInfo.put("issuer", jwt.getIssuer());
        tokenInfo.put("subject", jwt.getSubject());
        tokenInfo.put("audience", jwt.getAudience());
        tokenInfo.put("claims", jwt.getClaims());
        return ResponseEntity.ok(tokenInfo);
    }
}
