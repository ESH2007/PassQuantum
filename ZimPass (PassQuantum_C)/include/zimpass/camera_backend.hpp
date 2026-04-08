#pragma once

#include <memory>
#include <opencv2/opencv.hpp>

namespace zimpass {

class CameraBackend {
public:
    virtual ~CameraBackend() = default;

    virtual bool open(int device_index = 0) = 0;
    virtual cv::Mat read_frame() = 0;
    virtual void close() = 0;
};

std::unique_ptr<CameraBackend> create_camera_backend();

} // namespace zimpass
