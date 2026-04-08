#pragma once

#include "zimpass/types.hpp"

#include <array>
#include <string>
#include <vector>

namespace zimpass {

struct KeyMaterial {
    std::array<std::uint8_t, 32> encryption_key {};
    std::array<std::uint8_t, 32> verification_key {};
};

struct CipherEnvelope {
    std::vector<std::uint8_t> nonce;
    std::vector<std::uint8_t> ciphertext;
};

class CryptoEngine {
public:
    CryptoEngine();

    std::vector<std::uint8_t> random_bytes(std::size_t count) const;
    KdfParams default_kdf_params() const;
    KeyMaterial derive_keys(const std::string& master_password, KdfParams params) const;

    CipherEnvelope encrypt_aead(const std::vector<std::uint8_t>& plaintext,
                                const std::array<std::uint8_t, 32>& key) const;

    std::vector<std::uint8_t> decrypt_aead(const CipherEnvelope& envelope,
                                           const std::array<std::uint8_t, 32>& key) const;

    std::vector<std::uint8_t> hmac_sha256(const std::vector<std::uint8_t>& data,
                                          const std::array<std::uint8_t, 32>& key) const;

    std::vector<std::uint8_t> compute_private_key_fingerprint(const std::vector<std::uint8_t>& private_key_bytes) const;

    std::pair<std::vector<std::uint8_t>, std::vector<std::uint8_t>> kem_encapsulate(const std::vector<std::uint8_t>& public_key) const;
    std::vector<std::uint8_t> kem_decapsulate(const std::vector<std::uint8_t>& ciphertext,
                                              const std::vector<std::uint8_t>& secret_key) const;
};

} // namespace zimpass
