# PassQuantum Go Application Specification

## 1. Scope
This document specifies how the Go implementation in `new-passquantum` works, including runtime flow, security model, module boundaries, and a file-by-file description of all Go source files currently in the project (excluding vendor and tool caches).

## 2. System Summary
PassQuantum is a desktop password manager built with Fyne UI and a modular core:
- UI layer (`ui/`): windows, dialogs, navigation, password workflows, settings, biometric UX.
- Security core (`core/crypto/`): Argon2id KDF, AES-GCM encryption, HMAC integrity, Kyber768 KEM, app-security profile logic.
- Biometric core (`core/biometric/`): face-landmark pipeline (GoCV backend), template extraction, similarity checks, model path resolution, fallback stubs.
- Storage core (`core/storage/`): vault persistence and app security metadata persistence.
- Data model (`core/model/`): password entry binary format.
- CLI tools (`cmd/`, `.tmp_model_src/`): demos, diagnostics, and smoke tests.

## 3. Runtime Architecture

### 3.1 Main User Journey
1. App starts in `ui/main.go`.
2. Kyber keypair is loaded or generated (`core/crypto/kyber.go`).
3. Startup access state is resolved (`ui/access_control.go` + `core/storage/security_metadata.go`).
4. User either creates the master profile or unlocks existing profile (`core/crypto/app_security.go`).
5. User selects/creates vault, vault keys are derived per vault salt (`core/crypto/kdf.go`).
6. Vault is opened/decrypted (`core/storage/storage.go` + `core/crypto/vault.go`).
7. Password records are shown and managed via UI screens.
8. If biometric is enabled and available, continuous checks can lock session on repeated failures.

### 3.2 Security Layers
- App-level security profile (`app-security.pqmeta`):
  - Stores verifier derived from master password + private key fingerprint.
  - Stores biometric settings/template (format version 2).
- Vault-level encryption (`*.pqdb`):
  - Argon2id-derived encryption and verification keys.
  - AES-256-GCM for ciphertext confidentiality/integrity at cipher level.
  - HMAC-SHA256 over version + KDF metadata + encrypted payload for file integrity.
- Entry-level secret wrapping:
  - Each password uses Kyber encapsulation + AES-GCM payload encryption.

### 3.3 Build-Tag Split
- Biometric-enabled implementation: `//go:build !nobiometric && cgo`
- Stub/no-biometric implementation: `//go:build nobiometric || !cgo`
This allows the same UI and app logic to compile with or without GoCV/cgo.

## 4. Data Contracts

### 4.1 App Security Profile
Defined in `core/crypto/app_security.go`:
- `FormatVersion`
- `PrivateKeyFingerprint`
- `KDFParams`
- `Verifier`
- `Biometric` (enabled flag, threshold, optional camera index)
- `BiometricTemplate` (serialized feature vector)

### 4.2 Vault File
Defined in `core/crypto/vault.go`:
- Version
- Serialized KDF params
- HMAC
- Encrypted data (nonce + ciphertext)

### 4.3 Password Entry
Defined in `core/model/password_entry.go` and parsed in `core/storage/storage.go`:
- ID
- Service
- Username
- Kyber ciphertext
- AES nonce
- AES ciphertext

## 5. Component Behavior

### 5.1 App State and Session Handling
`ui/main.go` defines `AppState` with:
- Key material (public/private key, session keys, current vault keys)
- Unlock/session flags
- Current vault metadata
- Biometric runtime and template state
- Mutex protection (`sync.Mutex`) for sensitive mutable state

`ui/access_control.go` provides:
- startup profile detection and fingerprint mismatch handling
- profile creation and unlock verification
- vault creation/open under unlocked session
- master password rotation with staged atomic replacement (`.tmp` strategy)
- secure state cleanup and memory wipe paths

### 5.2 Password Storage Flow
- Serialize entries to plaintext (`serializeEntries`).
- Encrypt full vault (`EncryptVault`).
- Persist bytes to disk (`WriteVault`).
- Read path performs reverse parse/decrypt/entry reconstruction (`ReadVault`).

### 5.3 Biometric Flow
- `core/biometric/common.go` provides feature extraction, cosine similarity, serialization helpers, thresholds.
- `core/biometric/biometric.go` (cgo) runs model inference pipeline:
  - face detection/crop
  - face mesh landmark prediction
  - runtime checks and mesh drawing helpers
- `ui/biometric_control.go` integrates capture, verify, enrolment, and continuous check loop.
- Stub files return unavailable behavior when biometric backend is disabled.

### 5.4 UI Flow
Primary UI screens:
- Login/unlock setup: `ui/login_screen.go`
- Vault picker: `ui/vault_selection.go`
- Main shell/navigation: `ui/main_screen.go`
- Password list/generator: `ui/passwords_view.go`
- Password checker: `ui/password_checker.go`
- Settings and biometric enrolment: `ui/settings_screen.go`
Supporting visuals/components:
- custom dialogs, themes, enhanced widgets, color picker.

## 6. File-by-File Specification (All 40 Go Files)

### 6.1 Diagnostics and Temporary Model Tools (`.tmp_model_src/`)
1. `.tmp_model_src/camera_probe.go`
- Purpose: probe camera APIs and indexes.
- Key functions: `test`, `main`.
- Role: local troubleshooting utility, not required for main app runtime.

2. `.tmp_model_src/model_introspect.go`
- Purpose: inspect model metadata/compatibility details.
- Key function: `main`.
- Role: developer diagnostic tool.

3. `.tmp_model_src/model_smoketest.go`
- Purpose: run model load smoke tests.
- Key functions: `testModel`, `main`.
- Role: model validation utility.

### 6.2 Command Programs (`cmd/`)
4. `cmd/biometric-demo/main.go` (`!nobiometric && cgo`)
- Purpose: interactive biometric camera demo with overlay and matching status.
- Key functions: `main`, `shouldQuit`, `cloneFeatures`, `openCamera`, `drawOverlay`.
- Role: standalone biometric demonstration and tuning aid.

5. `cmd/biometric-demo/main_nocgo.go` (`!nobiometric && !cgo`)
- Purpose: fallback entrypoint when cgo is unavailable.
- Key function: `main`.
- Role: clearly reports unsupported biometric demo in no-cgo context.

6. `cmd/test-vault/main.go`
- Purpose: vault test runner utility.
- Key function: `main`.
- Role: manual validation of vault read/write flows.

### 6.3 Biometric Core (`core/biometric/`)
7. `core/biometric/common.go`
- Purpose: backend-neutral biometric constants and math utilities.
- Key symbols: `Landmark`, `RuntimeHandle`, `NewDefaultRuntime`, `ExtractFeatures`, `CosineSimilarity`, `SerializeFeatures`, `DeserializeFeatures`, `EffectiveThreshold`.
- Role: shared logic for both UI and backend pipeline.

8. `core/biometric/biometric.go` (`!nobiometric && cgo`)
- Purpose: concrete GoCV/ONNX runtime implementation.
- Key symbols: `FaceDetector`, `FaceMesh`, `Pipeline`, `NewFaceDetector`, `NewFaceMesh`, `NewPipeline`, `Run`, `RunFrame`, `BackendName`, `DrawMesh`.
- Role: production biometric inference pipeline.

9. `core/biometric/biometric_stub.go` (`nobiometric || !cgo`)
- Purpose: stub backend for builds without biometric backend support.
- Key symbols mirror production API.
- Role: compile-time compatibility and graceful feature disabling.

10. `core/biometric/model_paths.go`
- Purpose: resolve model files from env/executable/repository candidates; validate model content.
- Key functions: `ResolveDefaultModelPaths`, `resolveModelPath`, `candidateModelPaths`, `ValidateModelFile`.
- Role: robust model discovery and corrupted-file rejection.

11. `core/biometric/biometric_test.go`
- Purpose: test feature math, thresholds, serialization, path resolution, and validation logic.
- Role: regression safety for biometric utility layer.

### 6.4 Crypto Core (`core/crypto/`)
12. `core/crypto/kdf.go`
- Purpose: Argon2id key derivation and KDF metadata serialization.
- Key symbols: `KDFParams`, `DefaultKDFParams`, `GenerateSalt`, `DeriveKeys`, `WipeBytes`, `Serialize`, `KDFParamsDeserialize`.
- Role: root derivation for vault and profile security keys.

13. `core/crypto/vault.go`
- Purpose: encrypted vault container format and integrity verification.
- Key symbols: `VaultFile`, `EncryptVault`, `DecryptVault`, `Serialize`, `VaultFileDeserialize`.
- Role: authenticated at-rest vault persistence format.

14. `core/crypto/kyber.go`
- Purpose: Kyber768 keypair and encapsulation operations.
- Key functions: `GenerateKeypair`, `SaveKeypair`, `LoadKeypair`, `Encapsulate`, `Decapsulate`.
- Role: post-quantum secret exchange for entry-level encryption.

15. `core/crypto/aes.go`
- Purpose: AES-256-GCM encrypt/decrypt helpers for string payloads.
- Key functions: `EncryptAES256GCM`, `DecryptAES256GCM`.
- Role: symmetric encryption primitive wrapper.

16. `core/crypto/app_security.go`
- Purpose: app-level profile creation/verification, key fingerprint binding, biometric metadata schema.
- Key symbols: `AppSecurityFormatVersion`, `BiometricSettings`, `AppSecurityProfile`, `PrivateKeyFingerprint`, `CreateAppSecurityProfile`, `VerifyAppSecurityProfile`.
- Role: global app unlock gate.

17. `core/crypto/app_security_test.go`
- Purpose: profile creation/verification and mismatch tests.
- Role: validation of app-level security logic.

### 6.5 Model (`core/model/`)
18. `core/model/password_entry.go`
- Purpose: password entry in-memory model and binary serialization.
- Key symbols: `PasswordEntry`, `NewPasswordEntry`, `Serialize`, `Deserialize`.
- Role: atomic record unit stored inside vault plaintext before vault encryption.

### 6.6 Storage (`core/storage/`)
19. `core/storage/storage.go`
- Purpose: vault file write/read/exists/delete operations.
- Key symbols: `DefaultVaultFile`, `WriteVault`, `ReadVault`, `VaultExists`, `DeleteVault`.
- Role: gateway between encrypted vault bytes and decoded entries.

20. `core/storage/security_metadata.go`
- Purpose: app-security metadata persistence and vault re-encryption support.
- Key symbols: `DefaultAppSecurityMetadataPath`, `AppSecurityProfileExists`, `LoadAppSecurityProfile`, `SaveAppSecurityProfile`, `ReencryptVaultFile`.
- Role: lifecycle management for global unlock profile and master password rotation.

21. `core/storage/security_metadata_test.go`
- Purpose: tests for profile save/load, legacy compatibility, and vault re-encryption behavior.
- Role: regression tests for metadata and rotation paths.

### 6.7 UI Application (`ui/`)
22. `ui/main.go`
- Purpose: desktop app entrypoint, locale normalization, icon setup, key initialization.
- Key symbols: `AppState`, `main`, `normalizeLocaleForFyne`, `setApplicationIcon`, `initializeApp`.
- Role: bootstraps app runtime and first screen.

23. `ui/access_control.go`
- Purpose: unlock/setup orchestration, session key handling, master password changes, secure state clearing.
- Key symbols: `ResolveStartupAccessState`, `CreateMasterPasswordProfile`, `unlockAppSession`, `changeMasterPassword`, state store/clear methods.
- Role: central security workflow coordinator.

24. `ui/login_screen.go`
- Purpose: setup/unlock screens and optional biometric gate before full access.
- Key functions: `PromptMasterPassword`, `showCreateMasterPasswordScreen`, `showUnlockScreen`, `showFaceVerificationGate`.
- Role: user authentication UX.

25. `ui/vault_selection.go`
- Purpose: vault listing, creation, deletion dialogs.
- Key functions: `ShowVaultSelection`, `createVaultCard`, `showCreateVaultDialog`, `showDeleteVaultDialog`.
- Role: vault-level navigation hub.

26. `ui/main_screen.go`
- Purpose: top-level app shell and navigation state machine.
- Key symbols: `NavView`, `NavigationState`, `ShowMainScreen`, view constructors.
- Role: switches between password, vault, checker, generator, settings subviews.

27. `ui/passwords_view.go`
- Purpose: password list rendering, card actions, generator UI.
- Key symbols: `PasswordGeneratorSettings`, `ShowPasswordsView`, `displayPasswordsList`, `createPasswordCard`, `ShowPasswordGenerator`, `GeneratePassword`.
- Role: main day-to-day credential management screen.

28. `ui/password_checker.go`
- Purpose: password validation/checking interface.
- Key function: `ShowPasswordChecker`.
- Role: security quality checking UX.

29. `ui/settings_screen.go`
- Purpose: settings sections (security, biometric, vault, display, about) and palette customization.
- Key symbols: `SettingsSubview`, `ShowSettingsScreen`, `buildBiometricSettingsSection`, `showEnrolmentDialog`, `showChangeMasterPasswordDialog`.
- Role: app configuration and advanced controls.

30. `ui/helpers.go`
- Purpose: utility wrappers around storage/crypto operations and common vault actions.
- Key symbols: `GetVaultPath`, `ListVaults`, `CreateNewVault`, `UnlockVault`, `OpenVault`, wrappers for crypto calls, `ValidatePassword`.
- Role: convenience bridge utilities for UI workflows.

31. `ui/app_dialogs.go`
- Purpose: consistent information/warning/error/confirm dialogs.
- Key symbols: `ShowAppInformation`, `ShowAppWarning`, `ShowAppError`, `ShowAppConfirm`.
- Role: standardized app messaging and user prompts.

32. `ui/ui_theme.go`
- Purpose: theme constants, particle backgrounds, styled button/card/label factories.
- Key symbols: `Particle`, `ParticleBackground`, `GlowButton`, and component builder functions.
- Role: visual identity and reusable UI styling system.

33. `ui/ui_enhancements.go`
- Purpose: enhanced visual widgets (cards, toolbar, indicators, loaders, responsive containers).
- Key symbols: `CreateEnhancedCard`, `CreateStyledInput`, `CreateIconButton`, `CreateMetricCard`, etc.
- Role: richer component primitives used by screens.

34. `ui/color_picker.go`
- Purpose: custom hue/saturation/value picker widgets and gradient controls.
- Key symbols: `SVPicker`, `GradientSlider`, color conversion helpers.
- Role: display theme personalization.

35. `ui/biometric_camera_gocv.go` (`!nobiometric && cgo`)
- Purpose: camera selection and frame availability checks.
- Key symbols: `openBiometricCamera`, `candidateCameraIndices`, `cameraProducesFrames`.
- Role: reliable camera source selection for biometric flows.

36. `ui/camera_open.go` (`!nobiometric && cgo`)
- Purpose: Windows camera probing/opening helpers and warmup.
- Key symbols: `OpenCameraWindowsByName`, `OpenCamera`, `GetCameraDevicePaths`, `OpenCameraWindows`, `warmupCamera`.
- Role: low-level camera opening strategy for biometric runtime.

37. `ui/biometric_control.go` (`!nobiometric && cgo`)
- Purpose: biometric profile load/save, capture and verify, background continuous checking, lock-on-fail handling.
- Key symbols: `loadBiometricFromProfile`, `saveBiometricToProfile`, `ensureBiometricPipeline`, `captureAndVerifyFace`, `startContinuousCheck`, `runContinuousCheck`.
- Role: live biometric security control plane.

38. `ui/biometric_preview_gocv.go` (`!nobiometric && cgo`)
- Purpose: live enrolment preview stream for settings dialog.
- Key function: `startEnrollmentPreview`.
- Role: visual feedback during enrolment.

39. `ui/biometric_control_stub.go` (`nobiometric || !cgo`)
- Purpose: no-op/unsupported biometric behavior when backend unavailable.
- Role: keeps UI/security flow compile-safe without GoCV.

40. `ui/biometric_preview_stub.go` (`nobiometric || !cgo`)
- Purpose: preview stub for non-biometric builds.
- Role: interface compatibility.

## 7. Operational Notes
- Sensitive byte slices are actively wiped in multiple paths (`crypto.WipeBytes`).
- Vault and metadata updates during password change use temporary files and staged renames to reduce corruption risk.
- File permissions are explicitly restrictive for secret-bearing files where applicable.
- Biometric matching threshold defaults to `0.97` if no stored threshold is set.

## 8. Known Separation for Future C Port
The Go codebase maps cleanly to C modules:
- `crypto` -> `src/crypto`
- `storage` -> `src/storage`
- `model` -> `src/model`
- `biometric` -> `src/biometric`
- `ui` -> `src/ui`
- `cmd tools` -> `apps/`
This mapping is used to initialize `ZimPass (PassQuantum_C)`.
