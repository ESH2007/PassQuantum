#include "zimpass/crypto_engine.hpp"

#include <oqs/oqs.h>
#include <sodium.h>

#include <stdexcept>
#include <string_view>

namespace zimpass {

namespace {

std::string kem_algorithm() {
#ifdef OQS_KEM_alg_ml_kem_768
    return OQS_KEM_alg_ml_kem_768;
#elif defined(OQS_KEM_alg_kyber_768)
    return OQS_KEM_alg_kyber_768;
#else
    throw std::runtime_error("No supported ML-KEM/Kyber algorithm found in liboqs build");
#endif
}

std::array<std::uint8_t, 32> derive_domain_key(const std::vector<std::uint8_t>& master_key, std::string_view domain) {
    std::array<std::uint8_t, 32> out {};
    crypto_generichash_state state {};
    crypto_generichash_init(&state, nullptr, 0, out.size());
    crypto_generichash_update(&state, reinterpret_cast<const unsigned char*>(domain.data()), domain.size());
    crypto_generichash_update(&state, master_key.data(), master_key.size());
    crypto_generichash_final(&state, out.data(), out.size());
    return out;
}

} // namespace

CryptoEngine::CryptoEngine() {
    if (sodium_init() < 0) {
        throw std::runtime_error("libsodium initialization failed");
    }
}

std::vector<std::uint8_t> CryptoEngine::random_bytes(std::size_t count) const {
    std::vector<std::uint8_t> out(count);
    randombytes_buf(out.data(), out.size());
    return out;
}

KdfParams CryptoEngine::default_kdf_params() const {
    return KdfParams {};
}

KeyMaterial CryptoEngine::derive_keys(const std::string& master_password, KdfParams params) const {
    if (params.salt.empty()) {
        params.salt = random_bytes(16);
    }

    std::vector<std::uint8_t> master_key(64);
    if (crypto_pwhash(master_key.data(),
                      master_key.size(),
                      master_password.c_str(),
                      master_password.size(),
                      params.salt.data(),
                      params.iterations,
                      static_cast<size_t>(params.memory_kib) * 1024u,
                      crypto_pwhash_ALG_ARGON2ID13) != 0) {
        throw std::runtime_error("crypto_pwhash failed");
    }

    KeyMaterial km;
    km.encryption_key = derive_domain_key(master_key, "encryption");
    km.verification_key = derive_domain_key(master_key, "verification");

    sodium_memzero(master_key.data(), master_key.size());
    return km;
}

CipherEnvelope CryptoEngine::encrypt_aead(const std::vector<std::uint8_t>& plaintext,
                                          const std::array<std::uint8_t, 32>& key) const {
    CipherEnvelope env;
    env.nonce = random_bytes(crypto_aead_chacha20poly1305_ietf_NPUBBYTES);

    env.ciphertext.resize(plaintext.size() + crypto_aead_chacha20poly1305_ietf_ABYTES);
    unsigned long long cipher_len = 0;

    crypto_aead_chacha20poly1305_ietf_encrypt(
        env.ciphertext.data(),
        &cipher_len,
        plaintext.data(),
        plaintext.size(),
        nullptr,
        0,
        nullptr,
        env.nonce.data(),
        key.data());

    env.ciphertext.resize(static_cast<std::size_t>(cipher_len));
    return env;
}

std::vector<std::uint8_t> CryptoEngine::decrypt_aead(const CipherEnvelope& envelope,
                                                     const std::array<std::uint8_t, 32>& key) const {
    std::vector<std::uint8_t> plaintext(envelope.ciphertext.size());
    unsigned long long plain_len = 0;

    if (crypto_aead_chacha20poly1305_ietf_decrypt(
            plaintext.data(),
            &plain_len,
            nullptr,
            envelope.ciphertext.data(),
            envelope.ciphertext.size(),
            nullptr,
            0,
            envelope.nonce.data(),
            key.data()) != 0) {
        throw std::runtime_error("AEAD decrypt failed");
    }

    plaintext.resize(static_cast<std::size_t>(plain_len));
    return plaintext;
}

std::vector<std::uint8_t> CryptoEngine::hmac_sha256(const std::vector<std::uint8_t>& data,
                                                    const std::array<std::uint8_t, 32>& key) const {
    std::vector<std::uint8_t> digest(crypto_auth_hmacsha256_BYTES);
    crypto_auth_hmacsha256_state state {};
    crypto_auth_hmacsha256_init(&state, key.data(), key.size());
    crypto_auth_hmacsha256_update(&state, data.data(), data.size());
    crypto_auth_hmacsha256_final(&state, digest.data());
    return digest;
}

std::vector<std::uint8_t> CryptoEngine::compute_private_key_fingerprint(const std::vector<std::uint8_t>& private_key_bytes) const {
    std::vector<std::uint8_t> digest(32);
    crypto_generichash(digest.data(), digest.size(), private_key_bytes.data(), private_key_bytes.size(), nullptr, 0);
    return digest;
}

std::pair<std::vector<std::uint8_t>, std::vector<std::uint8_t>> CryptoEngine::kem_encapsulate(const std::vector<std::uint8_t>& public_key) const {
    OQS_KEM* kem = OQS_KEM_new(kem_algorithm().c_str());
    if (!kem) {
        throw std::runtime_error("Failed to create OQS KEM");
    }

    std::vector<std::uint8_t> ciphertext(kem->length_ciphertext);
    std::vector<std::uint8_t> shared_secret(kem->length_shared_secret);

    if (OQS_KEM_encaps(kem, ciphertext.data(), shared_secret.data(), public_key.data()) != OQS_SUCCESS) {
        OQS_KEM_free(kem);
        throw std::runtime_error("KEM encapsulation failed");
    }

    OQS_KEM_free(kem);
    return {ciphertext, shared_secret};
}

std::vector<std::uint8_t> CryptoEngine::kem_decapsulate(const std::vector<std::uint8_t>& ciphertext,
                                                        const std::vector<std::uint8_t>& secret_key) const {
    OQS_KEM* kem = OQS_KEM_new(kem_algorithm().c_str());
    if (!kem) {
        throw std::runtime_error("Failed to create OQS KEM");
    }

    std::vector<std::uint8_t> shared_secret(kem->length_shared_secret);

    if (OQS_KEM_decaps(kem, shared_secret.data(), ciphertext.data(), secret_key.data()) != OQS_SUCCESS) {
        OQS_KEM_free(kem);
        throw std::runtime_error("KEM decapsulation failed");
    }

    OQS_KEM_free(kem);
    return shared_secret;
}

} // namespace zimpass
