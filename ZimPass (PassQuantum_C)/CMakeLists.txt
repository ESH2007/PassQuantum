cmake_minimum_required(VERSION 3.22)
project(ZimPass LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

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
