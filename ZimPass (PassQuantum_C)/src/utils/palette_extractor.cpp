#include "zimpass/palette_extractor.hpp"

#define STB_IMAGE_IMPLEMENTATION
#include <stb_image.h>

#include <algorithm>
#include <cstdint>
#include <stdexcept>
#include <unordered_map>

namespace zimpass {

namespace {

std::uint32_t quantize_rgba(unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    const std::uint32_t rq = static_cast<std::uint32_t>(r >> 4);
    const std::uint32_t gq = static_cast<std::uint32_t>(g >> 4);
    const std::uint32_t bq = static_cast<std::uint32_t>(b >> 4);
    const std::uint32_t aq = static_cast<std::uint32_t>(a >> 5);
    return (rq << 15) | (gq << 11) | (bq << 7) | aq;
}

Rgba dequantize_rgba(std::uint32_t key) {
    Rgba color {};
    color[0] = static_cast<unsigned char>(((key >> 15) & 0x0F) << 4);
    color[1] = static_cast<unsigned char>(((key >> 11) & 0x0F) << 4);
    color[2] = static_cast<unsigned char>(((key >> 7) & 0x0F) << 4);
    color[3] = static_cast<unsigned char>((key & 0x07) << 5);
    return color;
}

} // namespace

std::vector<Rgba> extract_top_palette_colors(const std::string& image_path, std::size_t count) {
    int width = 0;
    int height = 0;
    int channels = 0;

    unsigned char* pixels = stbi_load(image_path.c_str(), &width, &height, &channels, 4);
    if (!pixels) {
        throw std::runtime_error("stb_image failed to load image");
    }

    std::unordered_map<std::uint32_t, std::size_t> histogram;
    const std::size_t total = static_cast<std::size_t>(width) * static_cast<std::size_t>(height);

    for (std::size_t i = 0; i < total; ++i) {
        const auto* px = pixels + (i * 4);
        const auto key = quantize_rgba(px[0], px[1], px[2], px[3]);
        histogram[key]++;
    }

    stbi_image_free(pixels);

    std::vector<std::pair<std::uint32_t, std::size_t>> ranked(histogram.begin(), histogram.end());
    std::sort(ranked.begin(), ranked.end(), [](const auto& a, const auto& b) {
        return a.second > b.second;
    });

    if (count > ranked.size()) {
        count = ranked.size();
    }

    std::vector<Rgba> out;
    out.reserve(count);
    for (std::size_t i = 0; i < count; ++i) {
        out.push_back(dequantize_rgba(ranked[i].first));
    }

    return out;
}

} // namespace zimpass
