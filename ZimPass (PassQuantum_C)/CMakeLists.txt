cmake_minimum_required(VERSION 3.22)

if(CMAKE_HOST_SYSTEM_NAME STREQUAL "Linux")
    find_program(PASSQUANTUM_GCC NAMES gcc HINTS /usr/bin /home/linuxbrew/.linuxbrew/bin)
    find_program(PASSQUANTUM_GXX NAMES g++ HINTS /usr/bin /home/linuxbrew/.linuxbrew/bin)

    if(NOT DEFINED CMAKE_C_COMPILER)
        if(PASSQUANTUM_GCC)
            set(CMAKE_C_COMPILER "${PASSQUANTUM_GCC}" CACHE FILEPATH "C compiler" FORCE)
        endif()
    endif()

    if(NOT DEFINED CMAKE_CXX_COMPILER)
        if(PASSQUANTUM_GXX)
            set(CMAKE_CXX_COMPILER "${PASSQUANTUM_GXX}" CACHE FILEPATH "CXX compiler" FORCE)
        endif()
    endif()

    if(PASSQUANTUM_GCC)
        set(ENV{CC} "${PASSQUANTUM_GCC}")
    endif()
    if(PASSQUANTUM_GXX)
        set(ENV{CXX} "${PASSQUANTUM_GXX}")
    endif()
endif()

option(PASSQUANTUM_AUTO_INSTALL_DEPS "Automatically bootstrap vcpkg and install missing dependencies" ON)
set(PASSQUANTUM_VCPKG_ROOT "" CACHE PATH "Path to vcpkg root. Empty means auto-select a safe default.")

if(PASSQUANTUM_AUTO_INSTALL_DEPS AND NOT DEFINED CMAKE_TOOLCHAIN_FILE AND EXISTS "${CMAKE_CURRENT_LIST_DIR}/vcpkg.json")
    if(PASSQUANTUM_VCPKG_ROOT STREQUAL "")
        if(CMAKE_HOST_SYSTEM_NAME STREQUAL "Windows")
            set(PASSQUANTUM_VCPKG_ROOT "${CMAKE_CURRENT_LIST_DIR}/vcpkg" CACHE PATH "Path to vcpkg root. Empty means auto-select a safe default." FORCE)
        else()
            set(PASSQUANTUM_VCPKG_ROOT "$ENV{HOME}/.cache/passquantum-vcpkg" CACHE PATH "Path to vcpkg root. Empty means auto-select a safe default." FORCE)
        endif()
    endif()

    file(TO_CMAKE_PATH "${PASSQUANTUM_VCPKG_ROOT}" PASSQUANTUM_VCPKG_ROOT)

    if(NOT EXISTS "${PASSQUANTUM_VCPKG_ROOT}/.git")
        message(STATUS "vcpkg not found. Cloning into ${PASSQUANTUM_VCPKG_ROOT}")
        execute_process(
            COMMAND git clone https://github.com/microsoft/vcpkg.git "${PASSQUANTUM_VCPKG_ROOT}"
            RESULT_VARIABLE PASSQUANTUM_VCPKG_CLONE_RESULT
        )
        if(NOT PASSQUANTUM_VCPKG_CLONE_RESULT EQUAL 0)
            message(FATAL_ERROR "Failed to clone vcpkg. Install dependencies manually or set CMAKE_TOOLCHAIN_FILE.")
        endif()
    endif()

    if(WIN32)
        execute_process(
            COMMAND "${PASSQUANTUM_VCPKG_ROOT}/bootstrap-vcpkg.bat"
            WORKING_DIRECTORY "${PASSQUANTUM_VCPKG_ROOT}"
            RESULT_VARIABLE PASSQUANTUM_VCPKG_BOOTSTRAP_RESULT
        )
    else()
        execute_process(
            COMMAND "${PASSQUANTUM_VCPKG_ROOT}/bootstrap-vcpkg.sh"
            WORKING_DIRECTORY "${PASSQUANTUM_VCPKG_ROOT}"
            RESULT_VARIABLE PASSQUANTUM_VCPKG_BOOTSTRAP_RESULT
        )
    endif()

    if(NOT PASSQUANTUM_VCPKG_BOOTSTRAP_RESULT EQUAL 0)
        message(FATAL_ERROR "Failed to bootstrap vcpkg. Install dependencies manually or set CMAKE_TOOLCHAIN_FILE.")
    endif()

    set(CMAKE_TOOLCHAIN_FILE "${PASSQUANTUM_VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake" CACHE STRING "vcpkg toolchain" FORCE)
    set(VCPKG_FEATURE_FLAGS "manifests,binarycaching" CACHE STRING "vcpkg feature flags" FORCE)
    if(CMAKE_HOST_SYSTEM_NAME STREQUAL "Linux")
        set(VCPKG_INSTALL_OPTIONS "--allow-unsupported" CACHE STRING "vcpkg install options" FORCE)
        set(VCPKG_INSTALLED_DIR "$ENV{HOME}/.cache/passquantum-vcpkg-installed" CACHE PATH "vcpkg installed directory" FORCE)
    endif()
    message(STATUS "Using vcpkg toolchain: ${CMAKE_TOOLCHAIN_FILE}")
endif()

project(ZimPass LANGUAGES CXX)

if(CMAKE_SYSTEM_NAME STREQUAL "Linux" AND NOT CMAKE_CXX_COMPILER_ID STREQUAL "GNU")
    message(FATAL_ERROR "Linux builds must use GCC/G++. Reconfigure with -DCMAKE_C_COMPILER=gcc -DCMAKE_CXX_COMPILER=g++.")
endif()

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

set(PASSQUANTUM_TARGET_PLATFORM "AUTO" CACHE STRING "Target platform macro: AUTO, WINDOWS, LINUX, MACOS")
set_property(CACHE PASSQUANTUM_TARGET_PLATFORM PROPERTY STRINGS AUTO WINDOWS LINUX MACOS)
string(TOUPPER "${PASSQUANTUM_TARGET_PLATFORM}" PASSQUANTUM_TARGET_PLATFORM_UPPER)

if(PASSQUANTUM_TARGET_PLATFORM_UPPER STREQUAL "AUTO")
    if(CMAKE_SYSTEM_NAME STREQUAL "Windows")
        set(PASSQUANTUM_TARGET_PLATFORM_UPPER "WINDOWS")
    elseif(CMAKE_SYSTEM_NAME STREQUAL "Linux")
        set(PASSQUANTUM_TARGET_PLATFORM_UPPER "LINUX")
    elseif(CMAKE_SYSTEM_NAME STREQUAL "Darwin")
        set(PASSQUANTUM_TARGET_PLATFORM_UPPER "MACOS")
    else()
        message(FATAL_ERROR "Unsupported platform for PASSQUANTUM_TARGET_PLATFORM=AUTO: ${CMAKE_SYSTEM_NAME}")
    endif()
endif()

if(PASSQUANTUM_TARGET_PLATFORM_UPPER STREQUAL "WINDOWS")
    add_compile_definitions(PASSQUANTUM_WINDOWS=1)
elseif(PASSQUANTUM_TARGET_PLATFORM_UPPER STREQUAL "LINUX")
    add_compile_definitions(PASSQUANTUM_LINUX=1)
elseif(PASSQUANTUM_TARGET_PLATFORM_UPPER STREQUAL "MACOS")
    add_compile_definitions(PASSQUANTUM_MACOS=1)
else()
    message(FATAL_ERROR "Invalid PASSQUANTUM_TARGET_PLATFORM value: ${PASSQUANTUM_TARGET_PLATFORM}")
endif()

option(ZIMPASS_ENABLE_BIOMETRIC "Enable OpenCV+dlib biometric module" ON)
option(ZIMPASS_ENABLE_GUI "Enable Dear ImGui + SDL2 desktop UI" ON)

find_path(NLOHMANN_JSON_INCLUDE_DIR nlohmann/json.hpp)
if(NOT NLOHMANN_JSON_INCLUDE_DIR)
    message(WARNING "nlohmann/json not found. Metadata JSON support will use fallback serializer.")
endif()

find_package(SDL2 QUIET)
find_package(OpenCV QUIET)
find_package(dlib QUIET)

find_path(SODIUM_INCLUDE_DIR sodium.h)
find_library(SODIUM_LIBRARY sodium)
if(NOT SODIUM_INCLUDE_DIR OR NOT SODIUM_LIBRARY)
    message(FATAL_ERROR "libsodium not found")
endif()

find_path(OQS_INCLUDE_DIR oqs/oqs.h)
find_library(OQS_LIBRARY oqs)
if(NOT OQS_INCLUDE_DIR OR NOT OQS_LIBRARY)
    message(FATAL_ERROR "liboqs not found")
endif()

find_path(SQLCIPHER_INCLUDE_DIR sqlite3.h)
find_library(SQLCIPHER_LIBRARY sqlcipher)
if(NOT SQLCIPHER_INCLUDE_DIR OR NOT SQLCIPHER_LIBRARY)
    message(FATAL_ERROR "SQLCipher not found")
endif()

find_path(ZXCVBN_INCLUDE_DIR zxcvbn.h)
find_library(ZXCVBN_LIBRARY zxcvbn)
if(NOT ZXCVBN_INCLUDE_DIR OR NOT ZXCVBN_LIBRARY)
    message(WARNING "zxcvbn-c not found. Password strength will use fallback heuristic.")
endif()

find_path(IMGUI_INCLUDE_DIR imgui.h)
find_library(IMGUI_LIBRARY imgui)
if(ZIMPASS_ENABLE_GUI)
    if(NOT SDL2_FOUND OR NOT IMGUI_INCLUDE_DIR OR NOT IMGUI_LIBRARY)
        message(WARNING "GUI dependencies not fully found (SDL2 + Dear ImGui). Disabling GUI target.")
        set(ZIMPASS_ENABLE_GUI OFF)
    endif()
endif()

if(ZIMPASS_ENABLE_BIOMETRIC)
    if(NOT OpenCV_FOUND OR NOT TARGET dlib::dlib)
        message(WARNING "Biometric dependencies not fully found (OpenCV + dlib). Disabling biometric module.")
        set(ZIMPASS_ENABLE_BIOMETRIC OFF)
    endif()
endif()

add_library(zimpass_core
    src/security/crypto_engine.cpp
    src/vault/vault_repository.cpp
    src/utils/password_strength.cpp
    src/utils/palette_extractor.cpp
    src/biometric/camera_backend.cpp
    src/biometric/face_engine.cpp
    src/app/application.cpp
)

target_include_directories(zimpass_core
    PUBLIC include
    PRIVATE
        ${SODIUM_INCLUDE_DIR}
        ${OQS_INCLUDE_DIR}
        ${SQLCIPHER_INCLUDE_DIR}
        ${NLOHMANN_JSON_INCLUDE_DIR}
        ${ZXCVBN_INCLUDE_DIR}
)

target_link_libraries(zimpass_core
    PUBLIC
    PRIVATE ${SODIUM_LIBRARY} ${OQS_LIBRARY} ${SQLCIPHER_LIBRARY}
)

if(NLOHMANN_JSON_INCLUDE_DIR)
    target_compile_definitions(zimpass_core PRIVATE ZIMPASS_HAS_NLOHMANN_JSON=1)
endif()

if(ZXCVBN_INCLUDE_DIR AND ZXCVBN_LIBRARY)
    target_compile_definitions(zimpass_core PRIVATE ZIMPASS_HAS_ZXCVBN=1)
    target_link_libraries(zimpass_core PRIVATE ${ZXCVBN_LIBRARY})
endif()

if(ZIMPASS_ENABLE_BIOMETRIC)
    target_compile_definitions(zimpass_core PRIVATE ZIMPASS_ENABLE_BIOMETRIC=1)
    target_include_directories(zimpass_core PRIVATE ${OpenCV_INCLUDE_DIRS})
    target_link_libraries(zimpass_core PRIVATE ${OpenCV_LIBS} dlib::dlib)
endif()

if(ZIMPASS_ENABLE_GUI)
    add_executable(zimpass_desktop
        src/ui/main.cpp
        src/ui/gui_app.cpp
    )

    target_include_directories(zimpass_desktop PRIVATE ${IMGUI_INCLUDE_DIR})
    target_link_libraries(zimpass_desktop PRIVATE zimpass_core SDL2::SDL2 ${IMGUI_LIBRARY})
endif()
