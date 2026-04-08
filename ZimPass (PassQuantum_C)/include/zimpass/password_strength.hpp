#pragma once

#include <string>

namespace zimpass {

struct PasswordStrength {
    int score {0};
    std::string warning;
    std::string feedback;
};

PasswordStrength evaluate_password_strength(const std::string& password);

} // namespace zimpass
