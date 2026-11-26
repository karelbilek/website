FROM rust:1-slim-trixie AS build
WORKDIR /src
COPY Cargo.toml Cargo.lock ./
COPY src src
RUN cargo build --release

FROM debian:trixie-slim
WORKDIR /app
COPY --from=build /src/target/release/website ./website
COPY public public
COPY static static
COPY chronocomics/dist chronocomics/dist
# written by the GitHub workflow from secrets, gitignored
COPY gemini-cert.pem gemini-key.rsa ./
ENV RUST_LOG=info
EXPOSE 8080 1965
CMD ["/app/website"]
