/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  env: {
    GATEWAY_URL: process.env.GATEWAY_URL || 'http://localhost:8888',
  },
  output: 'standalone'
};

export default nextConfig;
