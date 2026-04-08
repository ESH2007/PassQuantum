#include "zimpass/application.hpp"

#include <stdexcept>

namespace zimpass {

Application::Application()
    : vault_(std::make_unique<VaultRepository>(crypto_)) {}

void Application::open_vault(const std::string& path, const std::string& passphrase) {
    vault_->open_or_create(path, passphrase);
    session_.current_vault = path;
}

void Application::lock() {
    session_.unlocked = false;
    session_.session_keys.reset();
}

bool Application::verify_master_password(const std::string& master_password) {
    const auto profile = vault_->read_app_security_profile();
    auto keys = crypto_.derive_keys(master_password, profile.kdf);

    std::vector<std::uint8_t> verifier_input;
    verifier_input.insert(verifier_input.end(), profile.private_key_fingerprint.begin(), profile.private_key_fingerprint.end());
    const auto computed = crypto_.hmac_sha256(verifier_input, keys.verification_key);

    const bool ok = computed == profile.verifier;
    if (ok) {
        session_.session_keys = keys;
        session_.unlocked = true;
    }

    return ok;
}

void Application::create_master_profile(const std::string& master_password, const std::vector<std::uint8_t>& private_key) {
    AppSecurityProfile profile;
    profile.format_version = 2;
    profile.private_key_fingerprint = crypto_.compute_private_key_fingerprint(private_key);
    profile.kdf = crypto_.default_kdf_params();
    if (profile.kdf.salt.empty()) {
        profile.kdf.salt = crypto_.random_bytes(16);
    }

    auto keys = crypto_.derive_keys(master_password, profile.kdf);

    std::vector<std::uint8_t> verifier_input;
    verifier_input.insert(verifier_input.end(), profile.private_key_fingerprint.begin(), profile.private_key_fingerprint.end());
    profile.verifier = crypto_.hmac_sha256(verifier_input, keys.verification_key);

    vault_->write_app_security_profile(profile);
    session_.session_keys = keys;
    session_.unlocked = true;
}

} // namespace zimpass
