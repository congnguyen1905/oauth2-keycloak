import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "OAuth2 Keycloak App",
  description: "Authentication with Keycloak, Auth Service and KrakenD Gateway",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
