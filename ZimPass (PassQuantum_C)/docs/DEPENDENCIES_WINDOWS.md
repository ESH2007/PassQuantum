# Windows Dependency Setup (vcpkg)

## Required libraries for current scaffold
- libsodium
- liboqs
- sqlcipher
- sdl2
- imgui
- opencv4
- dlib
- zxcvbn-c
- nlohmann-json

## Suggested install flow
1. Install vcpkg and integrate with Visual Studio.
2. Install dependencies:

```powershell
vcpkg install libsodium:x64-windows liboqs:x64-windows sqlcipher:x64-windows sdl2:x64-windows imgui:x64-windows opencv4:x64-windows dlib:x64-windows nlohmann-json:x64-windows zxcvbn-c:x64-windows
```

3. Configure CMake using vcpkg toolchain:

```powershell
cmake -S "ZimPass (PassQuantum_C)" -B "ZimPass (PassQuantum_C)/build" -DCMAKE_TOOLCHAIN_FILE="C:/path/to/vcpkg/scripts/buildsystems/vcpkg.cmake"
```

4. Build:

```powershell
cmake --build "ZimPass (PassQuantum_C)/build" --config Release
```

## Notes
- If `zxcvbn-c` is unavailable in your vcpkg registry, the scaffold falls back to a heuristic checker.
- If OpenCV/dlib or ImGui/SDL2 are missing, CMake auto-disables biometric or GUI targets respectively.
- libsodium, liboqs, and SQLCipher are currently mandatory for core target configuration.
