package com.example.oauth2demo.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.annotation.Order;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.annotation.method.configuration.EnableGlobalMethodSecurity;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.oauth2.core.oidc.user.OidcUser;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.security.oauth2.server.resource.authentication.JwtAuthenticationConverter;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.util.matcher.AntPathRequestMatcher;

import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Configuration
@EnableWebSecurity
@EnableGlobalMethodSecurity(prePostEnabled = true)
public class SecurityConfig {

    /**
     * Chain 1 (Order 1): Resource Server cho các request /api/**
     * - Stateless, xác thực bằng JWT Bearer token
     * - KHÔNG redirect sang login page
     */
    @Bean
    @Order(1)
    public SecurityFilterChain apiSecurityFilterChain(HttpSecurity http) throws Exception {
        http
            .requestMatcher(new AntPathRequestMatcher("/api/**"))
            .authorizeRequests(auth -> auth
                .antMatchers(HttpMethod.GET, "/api/public/**").permitAll()
                .antMatchers("/api/admin/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .oauth2ResourceServer(oauth2 -> oauth2
                .jwt(jwt -> jwt.jwtAuthenticationConverter(jwtAuthenticationConverter()))
            )
            .sessionManagement(session ->
                session.sessionCreationPolicy(SessionCreationPolicy.STATELESS)
            )
            .csrf().disable();

        return http.build();
    }

    /**
     * Chain 2 (Order 2): OAuth2 Client cho browser requests
     * - Session-based, redirect sang Keycloak login
     * - Dùng Authorization Code Flow (PKCE)
     */
    @Bean
    @Order(2)
    public SecurityFilterChain webSecurityFilterChain(HttpSecurity http) throws Exception {
        http
            .authorizeRequests(auth -> auth
                .antMatchers("/", "/login", "/logout", "/error", "/css/**", "/js/**").permitAll()
                .antMatchers("/admin/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .oauth2Login(oauth2 -> oauth2
                .loginPage("/login")
                .defaultSuccessUrl("/dashboard", true)
                .failureUrl("/login?error=true")
            )
            .logout(logout -> logout
                // Cho phép cả GET /logout (link thông thường), không chỉ POST
                .logoutRequestMatcher(new AntPathRequestMatcher("/logout", "GET"))
                .invalidateHttpSession(true)
                .clearAuthentication(true)
                .deleteCookies("JSESSIONID")
                // logoutSuccessHandler phải đặt SAU cùng, nó override logoutSuccessUrl
                .logoutSuccessHandler(keycloakLogoutHandler())
            );

        return http.build();
    }

    /**
     * Converter: đọc roles từ JWT claim "roles" (được map bởi Keycloak)
     */
    @Bean
    public JwtAuthenticationConverter jwtAuthenticationConverter() {
        JwtAuthenticationConverter converter = new JwtAuthenticationConverter();
        converter.setJwtGrantedAuthoritiesConverter(jwt -> extractRoles(jwt));
        return converter;
    }

    private Collection<GrantedAuthority> extractRoles(Jwt jwt) {
        // Keycloak đặt roles trong claim "roles" (theo mapper đã cấu hình)
        List<String> roles = jwt.getClaimAsStringList("roles");
        if (roles == null) {
            // Fallback: thử đọc từ realm_access.roles (mặc định của Keycloak)
            Map<String, Object> realmAccess = jwt.getClaimAsMap("realm_access");
            if (realmAccess != null) {
                Object rolesObj = realmAccess.get("roles");
                if (rolesObj instanceof List) {
                    roles = (List<String>) rolesObj;
                }
            }
        }
        if (roles == null) return Collections.emptyList();

        return roles.stream()
            .map(role -> new SimpleGrantedAuthority(
                role.startsWith("ROLE_") ? role : "ROLE_" + role
            ))
            .collect(Collectors.toList());
    }

    /**
     * Keycloak logout handler: sau khi logout local, redirect logout Keycloak
     */
    @Bean
    public KeycloakLogoutHandler keycloakLogoutHandler() {
        return new KeycloakLogoutHandler();
    }
}