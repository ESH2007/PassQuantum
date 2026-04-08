#include "zimpass/password_strength.hpp"

#ifdef ZIMPASS_HAS_ZXCVBN
#include <zxcvbn.h>
#endif

namespace zimpass {

PasswordStrength evaluate_password_strength(const std::string& password) {
#ifdef ZIMPASS_HAS_ZXCVBN
    const zxcvbn_result_t result = zxcvbn_password_strength(password.c_str(), nullptr, 0);

    PasswordStrength out;
    out.score = result.score;
    if (result.feedback.warning) {
        out.warning = result.feedback.warning;
    }
    if (result.feedback.suggestions && result.feedback.suggestions[0]) {
        out.feedback = result.feedback.suggestions[0];
    }
    return out;
#else
    PasswordStrength out;
    const bool has_upper = password.find_first_of("ABCDEFGHIJKLMNOPQRSTUVWXYZ") != std::string::npos;
    const bool has_lower = password.find_first_of("abcdefghijklmnopqrstuvwxyz") != std::string::npos;
    const bool has_digit = password.find_first_of("0123456789") != std::string::npos;
    const bool has_symbol = password.find_first_of("!@#$%^&*()-_=+[]{};:'\",.<>?/\\|`") != std::string::npos;

    int score = 0;
    score += password.size() >= 12 ? 2 : 0;
    score += has_upper ? 1 : 0;
    score += has_lower ? 1 : 0;
    score += has_digit ? 1 : 0;
    score += has_symbol ? 1 : 0;
    if (score > 4) {
        score = 4;
    }

    out.score = score;
    out.warning = score < 3 ? "Weak password. Enable mixed character classes and length >= 12." : "";
    out.feedback = score < 4 ? "Consider adding random symbols and increasing length." : "Strong.";
    return out;
#endif
}

} // namespace zimpass
