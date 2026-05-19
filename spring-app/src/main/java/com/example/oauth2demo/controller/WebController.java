package com.example.oauth2demo.controller;

import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.core.oidc.user.OidcUser;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;

@Controller
public class WebController {

    @GetMapping("/")
    public String home() {
        return "home";
    }

    @GetMapping("/login")
    public String login() {
        return "login";
    }

    /**
     * Dashboard - yêu cầu đăng nhập
     * OidcUser chứa toàn bộ thông tin từ Keycloak (claims, tokens)
     */
    @GetMapping("/dashboard")
    public String dashboard(@AuthenticationPrincipal OidcUser oidcUser, Model model) {
        model.addAttribute("username", oidcUser.getPreferredUsername());
        model.addAttribute("email", oidcUser.getEmail());
        model.addAttribute("fullName", oidcUser.getFullName());
        model.addAttribute("roles", oidcUser.getAuthorities());
        model.addAttribute("accessToken", oidcUser.getIdToken().getTokenValue());
        model.addAttribute("claims", oidcUser.getClaims());
        return "dashboard";
    }

    /**
     * Admin page - chỉ ROLE_ADMIN
     */
    @GetMapping("/admin")
    public String admin(@AuthenticationPrincipal OidcUser oidcUser, Model model) {
        model.addAttribute("username", oidcUser.getPreferredUsername());
        return "admin";
    }
}
