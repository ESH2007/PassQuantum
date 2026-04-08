#pragma once

#include <array>
#include <string>
#include <vector>

namespace zimpass {

using Rgba = std::array<unsigned char, 4>;

std::vector<Rgba> extract_top_palette_colors(const std::string& image_path, std::size_t count);

} // namespace zimpass
