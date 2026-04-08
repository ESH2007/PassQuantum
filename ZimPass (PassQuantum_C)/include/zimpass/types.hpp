#pragma once

#include <cstdint>
#include <optional>
#include <string>
#include <vector>

namespace zimpass {

struct KdfParams {
    std::vector<std::uint8_t> salt;
    std::uint32_t memory_kib {64u * 1024u};
    std::uint32_t iterations {1u};
    std::uint8_t parallelism {4u};
};

struct BiometricSettings {
    bool enabled {false};
    float threshold {0.97f};
    std::optional<int> camera_index;
};

struct AppSecurityProfile {
    std::uint8_t format_version {2};
    std::vector<std::uint8_t> private_key_fingerprint;
    KdfParams kdf;
    std::vector<std::uint8_t> verifier;
    BiometricSettings biometric;
    std::vector<std::uint8_t> biometric_template;
};

struct PasswordEntry {
    std::uint64_t id {0};
    std::string service;
    std::string username;
    std::vector<std::uint8_t> kyber_ciphertext;
    std::vector<std::uint8_t> nonce;
    std::vector<std::uint8_t> ciphertext;
};

struct VaultRecord {
    std::vector<PasswordEntry> entries;
    KdfParams kdf;
};

} // namespace zimpass
