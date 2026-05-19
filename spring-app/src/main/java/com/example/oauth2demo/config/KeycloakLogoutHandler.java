package com.example.oauth2demo.config;

import org.springframework.security.core.Authentication;
import org.springframework.security.oauth2.core.oidc.user.OidcUser;
import org.springframework.security.web.authentication.logout.LogoutSuccessHandler;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

/**
 * Sau khi logout local session, redirect user tới Keycloak logout endpoint
 * để xóa SSO session trên Keycloak luôn.
 */
public class KeycloakLogoutHandler implements LogoutSuccessHandler {

    private static final String KEYCLOAK_LOGOUT_URL =
        "http://localhost:8080/realms/demo-realm/protocol/openid-connect/logout";

    @Override
    public void onLogoutSuccess(HttpServletRequest request,
                                 HttpServletResponse response,
                                 Authentication authentication) throws IOException {

        String redirectUrl = KEYCLOAK_LOGOUT_URL
            + "?post_logout_redirect_uri=http://localhost:8081/"
            + "&client_id=spring-boot-app";

        // Nếu có id_token, truyền theo để Keycloak biết ai đang logout
        if (authentication != null && authentication.getPrincipal() instanceof OidcUser) {
            OidcUser oidcUser = (OidcUser) authentication.getPrincipal();
            String idTokenValue = oidcUser.getIdToken().getTokenValue();
            redirectUrl += "&id_token_hint=" + idTokenValue;
        }

        response.sendRedirect(redirectUrl);
    }
}
