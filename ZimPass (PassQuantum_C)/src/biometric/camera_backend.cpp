#include "zimpass/camera_backend.hpp"

namespace zimpass {

class OpenCvCameraBackend final : public CameraBackend {
public:
    OpenCvCameraBackend() = default;
    ~OpenCvCameraBackend() override {
        close();
    }

    bool open(int device_index) override {
#if defined(PASSQUANTUM_WINDOWS)
        return capture_.open(device_index, cv::CAP_DSHOW);
#elif defined(PASSQUANTUM_LINUX)
        return capture_.open(device_index, cv::CAP_V4L2);
#elif defined(PASSQUANTUM_MACOS)
        return capture_.open(device_index, cv::CAP_AVFOUNDATION);
#else
#error "Unsupported platform macro. Define PASSQUANTUM_WINDOWS, PASSQUANTUM_LINUX, or PASSQUANTUM_MACOS."
#endif
    }

    cv::Mat read_frame() override {
        cv::Mat frame;
        if (!capture_.isOpened()) {
            return frame;
        }

        capture_ >> frame;
        return frame;
    }

    void close() override {
        if (capture_.isOpened()) {
            capture_.release();
        }
    }

private:
    cv::VideoCapture capture_;
};

std::unique_ptr<CameraBackend> create_camera_backend() {
    return std::make_unique<OpenCvCameraBackend>();
}

} // namespace zimpass
