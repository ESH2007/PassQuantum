Package: openssl:x64-linux@3.6.1#3

**Host Environment**

- Host: x64-linux
- Compiler: GNU 13.3.0
- CMake Version: 3.31.10
-    vcpkg-tool version: 2026-03-04-4b3e4c276b5b87a649e66341e11553e8c577459c
    vcpkg-scripts version: 81e3c888ef 2026-04-06 (26 hours ago)

**To Reproduce**

`vcpkg install --allow-unsupported`

**Failure logs**

```
Downloading https://github.com/openssl/openssl/archive/openssl-3.6.1.tar.gz -> openssl-openssl-openssl-3.6.1.tar.gz
Successfully downloaded openssl-openssl-openssl-3.6.1.tar.gz
-- Extracting source /home/lenovo/.cache/passquantum-vcpkg/downloads/openssl-openssl-openssl-3.6.1.tar.gz
-- Applying patch cmake-config.patch
-- Applying patch command-line-length.patch
-- Applying patch script-prefix.patch
-- Applying patch windows/install-layout.patch
-- Applying patch windows/install-pdbs.patch
-- Applying patch windows/install-programs.diff
-- Applying patch unix/android-cc.patch
-- Applying patch unix/move-openssldir.patch
-- Applying patch unix/no-empty-dirs.patch
-- Applying patch unix/no-static-libs-for-shared.patch
-- Applying patch fix-mingw-build.patch
-- Using source at /home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/src/nssl-3.6.1-444317d7e5.clean
-- Getting CMake variables for x64-linux
-- Loading CMake variables from /home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/cmake-get-vars_C_CXX-x64-linux.cmake.log
openssl requires Linux kernel headers from the system package manager.
   They can be installed on Alpine systems via `apk add linux-headers`.
   They can be installed on Ubuntu systems via `apt install linux-libc-dev`.

-- Getting CMake variables for x64-linux-dbg
-- Getting CMake variables for x64-linux-rel
-- Warning: Paths with embedded space may be handled incorrectly by configure:
   /home/lenovo/dev/PassQuantum/ZimPass (PassQuantum_C)/build/vcpkg_installed/x64-linux
   Please move the path to one without whitespaces!
-- Configuring x64-linux-dbg
CMake Error at scripts/cmake/vcpkg_execute_required_process.cmake:127 (message):
    Command failed: /usr/bin/bash -c "V=1 ./../src/nssl-3.6.1-444317d7e5.clean/vcpkg/configure  \"/usr/bin/perl\" \"/home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/src/nssl-3.6.1-444317d7e5.clean/Configure\" \"linux-x86_64\" \"enable-static-engine\" \"enable-capieng\" \"no-tests\" \"no-docs\" \"enable-ec_nistp_64_gcc_128\" \"no-shared\" \"no-module\" \"no-apps\" \"--openssldir=/etc/ssl\" \"--libdir=lib\" \"--disable-silent-rules\" \"--verbose\" \"--disable-shared\" \"--enable-static\" \"--debug\" \"--prefix=/home/lenovo/dev/PassQuantum/ZimPass (PassQuantum_C)/build/vcpkg_installed/x64-linux/debug\""
    Working Directory: /home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/x64-linux-dbg
    Error code: 255
    See logs for more information:
      /home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/config-x64-linux-dbg-err.log

Call Stack (most recent call first):
  scripts/cmake/vcpkg_configure_make.cmake:867 (vcpkg_execute_required_process)
  buildtrees/versioning_/versions/openssl/4d02bbb044439adeced1af6fc8a45f819c8bb5e2/unix/portfile.cmake:127 (vcpkg_configure_make)
  buildtrees/versioning_/versions/openssl/4d02bbb044439adeced1af6fc8a45f819c8bb5e2/portfile.cmake:81 (include)
  scripts/ports.cmake:206 (include)



```

<details><summary>/home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/config-x64-linux-dbg-err.log</summary>

```
+ /usr/bin/perl /home/lenovo/.cache/passquantum-vcpkg/buildtrees/openssl/src/nssl-3.6.1-444317d7e5.clean/Configure linux-x86_64 enable-static-engine enable-capieng no-tests no-docs enable-ec_nistp_64_gcc_128 no-shared no-module no-apps --openssldir=/etc/ssl --libdir=lib --debug --prefix=/home/lenovo/dev/PassQuantum/ZimPass '(PassQuantum_C)/build/vcpkg_installed/x64-linux/debug'

Failure!  build file wasn't produced.
Please read INSTALL.md and associated NOTES-* files.  You may also have to
look over your available compiler tool chain or change your configuration.

target already defined - linux-x86_64 (offending arg: (PassQuantum_C)/build/vcpkg_installed/x64-linux/debug)
```
</details>

**Additional context**

<details><summary>vcpkg.json</summary>

```
{
  "name": "zimpass",
  "version-string": "0.1.0",
  "builtin-baseline": "81e3c888ef717a026eb57bb2f3cb0dbd30246271",
  "dependencies": [
    "libsodium",
    "liboqs",
    "sqlcipher",
    "nlohmann-json",
    "opencv",
    {
      "name": "dlib",
      "default-features": false,
      "platform": "windows | osx"
    },
    "sdl2",
    "imgui"
  ]
}

```
</details>
