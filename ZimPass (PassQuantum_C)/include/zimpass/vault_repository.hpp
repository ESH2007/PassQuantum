#pragma once

#include "zimpass/types.hpp"
#include "zimpass/crypto_engine.hpp"

#include <string>
#include <vector>

struct sqlite3;

namespace zimpass {

class VaultRepository {
public:
    explicit VaultRepository(CryptoEngine engine);

    void open_or_create(const std::string& db_path, const std::string& passphrase);
    void close();

    void write_app_security_profile(const AppSecurityProfile& profile);
    AppSecurityProfile read_app_security_profile() const;

    void upsert_password(const PasswordEntry& entry);
    std::vector<PasswordEntry> list_passwords() const;

private:
    void initialize_schema();
    static std::string profile_to_json(const AppSecurityProfile& profile);
    static AppSecurityProfile profile_from_json(const std::string& json_text);

    CryptoEngine crypto_;
    sqlite3* db_ {nullptr};
};

} // namespace zimpass
