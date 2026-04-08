#include "zimpass/vault_repository.hpp"

#ifdef ZIMPASS_HAS_NLOHMANN_JSON
#include <nlohmann/json.hpp>
#endif
#include <sqlite3.h>

#include <stdexcept>
#include <sstream>

namespace zimpass {

#ifdef ZIMPASS_HAS_NLOHMANN_JSON
using json = nlohmann::json;
#endif

namespace {

void run_sql(sqlite3* db, const char* sql) {
    char* err = nullptr;
    if (sqlite3_exec(db, sql, nullptr, nullptr, &err) != SQLITE_OK) {
        std::string msg = err ? err : "unknown SQLCipher error";
        sqlite3_free(err);
        throw std::runtime_error(msg);
    }
}

} // namespace

VaultRepository::VaultRepository(CryptoEngine engine)
    : crypto_(std::move(engine)) {}

void VaultRepository::open_or_create(const std::string& db_path, const std::string& passphrase) {
    close();

    if (sqlite3_open(db_path.c_str(), &db_) != SQLITE_OK) {
        throw std::runtime_error("Failed to open SQLCipher database");
    }

    std::string pragma_key = "PRAGMA key = '" + passphrase + "';";
    run_sql(db_, pragma_key.c_str());
    run_sql(db_, "PRAGMA cipher_page_size = 4096;");
    run_sql(db_, "PRAGMA kdf_iter = 256000;");
    initialize_schema();
}

void VaultRepository::close() {
    if (db_) {
        sqlite3_close(db_);
        db_ = nullptr;
    }
}

void VaultRepository::initialize_schema() {
    run_sql(db_, "CREATE TABLE IF NOT EXISTS app_metadata (id INTEGER PRIMARY KEY CHECK (id=1), profile_json TEXT NOT NULL);");
    run_sql(db_, "CREATE TABLE IF NOT EXISTS vault_entries (id INTEGER PRIMARY KEY, service TEXT NOT NULL, username TEXT NOT NULL, kyber_ct BLOB NOT NULL, nonce BLOB NOT NULL, ciphertext BLOB NOT NULL);");
}

void VaultRepository::write_app_security_profile(const AppSecurityProfile& profile) {
    const std::string payload = profile_to_json(profile);

    sqlite3_stmt* stmt = nullptr;
    const char* sql = "INSERT INTO app_metadata(id, profile_json) VALUES(1, ?) ON CONFLICT(id) DO UPDATE SET profile_json=excluded.profile_json;";
    if (sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr) != SQLITE_OK) {
        throw std::runtime_error("Failed to prepare metadata upsert");
    }

    sqlite3_bind_text(stmt, 1, payload.c_str(), -1, SQLITE_TRANSIENT);
    if (sqlite3_step(stmt) != SQLITE_DONE) {
        sqlite3_finalize(stmt);
        throw std::runtime_error("Failed to upsert metadata");
    }

    sqlite3_finalize(stmt);
}

AppSecurityProfile VaultRepository::read_app_security_profile() const {
    sqlite3_stmt* stmt = nullptr;
    const char* sql = "SELECT profile_json FROM app_metadata WHERE id = 1;";
    if (sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr) != SQLITE_OK) {
        throw std::runtime_error("Failed to prepare metadata select");
    }

    int rc = sqlite3_step(stmt);
    if (rc != SQLITE_ROW) {
        sqlite3_finalize(stmt);
        throw std::runtime_error("No app security profile found");
    }

    const unsigned char* txt = sqlite3_column_text(stmt, 0);
    std::string raw = txt ? reinterpret_cast<const char*>(txt) : "";
    sqlite3_finalize(stmt);
    return profile_from_json(raw);
}

void VaultRepository::upsert_password(const PasswordEntry& entry) {
    sqlite3_stmt* stmt = nullptr;
    const char* sql = "INSERT INTO vault_entries(id, service, username, kyber_ct, nonce, ciphertext) VALUES(?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET service=excluded.service, username=excluded.username, kyber_ct=excluded.kyber_ct, nonce=excluded.nonce, ciphertext=excluded.ciphertext;";

    if (sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr) != SQLITE_OK) {
        throw std::runtime_error("Failed to prepare vault upsert");
    }

    sqlite3_bind_int64(stmt, 1, static_cast<sqlite3_int64>(entry.id));
    sqlite3_bind_text(stmt, 2, entry.service.c_str(), -1, SQLITE_TRANSIENT);
    sqlite3_bind_text(stmt, 3, entry.username.c_str(), -1, SQLITE_TRANSIENT);
    sqlite3_bind_blob(stmt, 4, entry.kyber_ciphertext.data(), static_cast<int>(entry.kyber_ciphertext.size()), SQLITE_TRANSIENT);
    sqlite3_bind_blob(stmt, 5, entry.nonce.data(), static_cast<int>(entry.nonce.size()), SQLITE_TRANSIENT);
    sqlite3_bind_blob(stmt, 6, entry.ciphertext.data(), static_cast<int>(entry.ciphertext.size()), SQLITE_TRANSIENT);

    if (sqlite3_step(stmt) != SQLITE_DONE) {
        sqlite3_finalize(stmt);
        throw std::runtime_error("Failed to upsert vault entry");
    }

    sqlite3_finalize(stmt);
}

std::vector<PasswordEntry> VaultRepository::list_passwords() const {
    std::vector<PasswordEntry> out;

    sqlite3_stmt* stmt = nullptr;
    const char* sql = "SELECT id, service, username, kyber_ct, nonce, ciphertext FROM vault_entries ORDER BY service;";
    if (sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr) != SQLITE_OK) {
        throw std::runtime_error("Failed to prepare vault select");
    }

    while (sqlite3_step(stmt) == SQLITE_ROW) {
        PasswordEntry e;
        e.id = static_cast<std::uint64_t>(sqlite3_column_int64(stmt, 0));
        e.service = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 1));
        e.username = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 2));

        const auto* kyber = reinterpret_cast<const std::uint8_t*>(sqlite3_column_blob(stmt, 3));
        const int kyber_size = sqlite3_column_bytes(stmt, 3);
        e.kyber_ciphertext.assign(kyber, kyber + kyber_size);

        const auto* nonce = reinterpret_cast<const std::uint8_t*>(sqlite3_column_blob(stmt, 4));
        const int nonce_size = sqlite3_column_bytes(stmt, 4);
        e.nonce.assign(nonce, nonce + nonce_size);

        const auto* cipher = reinterpret_cast<const std::uint8_t*>(sqlite3_column_blob(stmt, 5));
        const int cipher_size = sqlite3_column_bytes(stmt, 5);
        e.ciphertext.assign(cipher, cipher + cipher_size);

        out.push_back(std::move(e));
    }

    sqlite3_finalize(stmt);
    return out;
}

std::string VaultRepository::profile_to_json(const AppSecurityProfile& profile) {
#ifdef ZIMPASS_HAS_NLOHMANN_JSON
    json j;
    j["format_version"] = profile.format_version;
    j["private_key_fingerprint"] = profile.private_key_fingerprint;
    j["kdf"]["salt"] = profile.kdf.salt;
    j["kdf"]["memory_kib"] = profile.kdf.memory_kib;
    j["kdf"]["iterations"] = profile.kdf.iterations;
    j["kdf"]["parallelism"] = profile.kdf.parallelism;
    j["verifier"] = profile.verifier;
    j["biometric"]["enabled"] = profile.biometric.enabled;
    j["biometric"]["threshold"] = profile.biometric.threshold;
    if (profile.biometric.camera_index.has_value()) {
        j["biometric"]["camera_index"] = profile.biometric.camera_index.value();
    }
    j["biometric_template"] = profile.biometric_template;
    return j.dump();
#else
    std::ostringstream oss;
    oss << "format_version=" << static_cast<int>(profile.format_version) << "\n";
    oss << "biometric_enabled=" << (profile.biometric.enabled ? 1 : 0) << "\n";
    oss << "biometric_threshold=" << profile.biometric.threshold << "\n";
    if (profile.biometric.camera_index.has_value()) {
        oss << "biometric_camera_index=" << profile.biometric.camera_index.value() << "\n";
    }
    return oss.str();
#endif
}

AppSecurityProfile VaultRepository::profile_from_json(const std::string& json_text) {
#ifdef ZIMPASS_HAS_NLOHMANN_JSON
    const json j = json::parse(json_text);
    AppSecurityProfile p;
    p.format_version = j.value("format_version", 2);
    p.private_key_fingerprint = j.value("private_key_fingerprint", std::vector<std::uint8_t>{});
    p.kdf.salt = j.at("kdf").value("salt", std::vector<std::uint8_t>{});
    p.kdf.memory_kib = j.at("kdf").value("memory_kib", 64u * 1024u);
    p.kdf.iterations = j.at("kdf").value("iterations", 1u);
    p.kdf.parallelism = j.at("kdf").value("parallelism", static_cast<std::uint8_t>(4));
    p.verifier = j.value("verifier", std::vector<std::uint8_t>{});
    p.biometric.enabled = j.at("biometric").value("enabled", false);
    p.biometric.threshold = j.at("biometric").value("threshold", 0.97f);
    if (j.at("biometric").contains("camera_index")) {
        p.biometric.camera_index = j.at("biometric").at("camera_index").get<int>();
    }
    p.biometric_template = j.value("biometric_template", std::vector<std::uint8_t>{});
    return p;
#else
    AppSecurityProfile p;
    std::istringstream iss(json_text);
    std::string line;
    while (std::getline(iss, line)) {
        const auto sep = line.find('=');
        if (sep == std::string::npos) {
            continue;
        }
        const std::string key = line.substr(0, sep);
        const std::string value = line.substr(sep + 1);
        if (key == "format_version") {
            p.format_version = static_cast<std::uint8_t>(std::stoi(value));
        } else if (key == "biometric_enabled") {
            p.biometric.enabled = (value == "1");
        } else if (key == "biometric_threshold") {
            p.biometric.threshold = std::stof(value);
        } else if (key == "biometric_camera_index") {
            p.biometric.camera_index = std::stoi(value);
        }
    }
    return p;
#endif
}

} // namespace zimpass
