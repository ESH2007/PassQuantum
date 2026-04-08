#pragma once

#include "zimpass/biometric_service.hpp"
#include "zimpass/crypto_engine.hpp"
#include "zimpass/types.hpp"
#include "zimpass/vault_repository.hpp"

#include <memory>
#include <optional>
#include <string>

namespace zimpass {

struct SessionState {
    bool unlocked {false};
    std::string current_vault;
    std::optional<KeyMaterial> session_keys;
};

class Application {
public:
    Application();

    void open_vault(const std::string& path, const std::string& passphrase);
    void lock();

    bool verify_master_password(const std::string& master_password);
    void create_master_profile(const std::string& master_password, const std::vector<std::uint8_t>& private_key);

    SessionState& session() { return session_; }
    VaultRepository& vault() { return *vault_; }

private:
    CryptoEngine crypto_;
    std::unique_ptr<VaultRepository> vault_;
    BiometricService biometric_;
    SessionState session_;
};

} // namespace zimpass
