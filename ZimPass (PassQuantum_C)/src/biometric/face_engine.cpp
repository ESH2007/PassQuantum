#include "zimpass/biometric_service.hpp"

#include <cmath>

#ifdef ZIMPASS_ENABLE_BIOMETRIC
#include <dlib/image_processing/frontal_face_detector.h>
#include <opencv2/opencv.hpp>
#endif

namespace zimpass {

bool BiometricService::initialize(const std::string&, const std::string&) {
#ifdef ZIMPASS_ENABLE_BIOMETRIC
    initialized_ = true;
#else
    initialized_ = false;
#endif
    return initialized_;
}

std::vector<float> BiometricService::extract_embedding_from_camera(int camera_index) {
#ifdef ZIMPASS_ENABLE_BIOMETRIC
    if (!initialized_) {
        return {};
    }

    cv::VideoCapture camera(camera_index);
    if (!camera.isOpened()) {
        return {};
    }

    cv::Mat frame;
    camera >> frame;
    if (frame.empty()) {
        return {};
    }

    cv::Mat gray;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);

    dlib::cv_image<unsigned char> dimg(gray);
    auto detector = dlib::get_frontal_face_detector();
    const auto faces = detector(dimg);
    if (faces.empty()) {
        return {};
    }

    // Placeholder embedding to establish the porting seam.
    // Next iteration: replace with landmark/face-net embedding pipeline.
    return {1.0f, 0.5f, 0.1f, 0.8f, 0.2f, 0.7f, 0.4f, 0.3f};
#else
    (void)camera_index;
    return {};
#endif
}

float BiometricService::cosine_similarity(const std::vector<float>& a, const std::vector<float>& b) const {
    if (a.size() != b.size() || a.empty()) {
        return 0.0f;
    }

    double dot = 0.0;
    double na = 0.0;
    double nb = 0.0;
    for (std::size_t i = 0; i < a.size(); ++i) {
        dot += static_cast<double>(a[i]) * static_cast<double>(b[i]);
        na += static_cast<double>(a[i]) * static_cast<double>(a[i]);
        nb += static_cast<double>(b[i]) * static_cast<double>(b[i]);
    }

    const double denom = std::sqrt(na) * std::sqrt(nb);
    if (denom <= 0.0) {
        return 0.0f;
    }
    return static_cast<float>(dot / denom);
}

} // namespace zimpass
