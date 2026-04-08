#pragma once

#include <string>
#include <vector>

namespace zimpass {

class BiometricService {
public:
    bool initialize(const std::string& face_detector_model, const std::string& landmark_model);
    std::vector<float> extract_embedding_from_camera(int camera_index);
    float cosine_similarity(const std::vector<float>& a, const std::vector<float>& b) const;

private:
    bool initialized_ {false};
};

} // namespace zimpass
