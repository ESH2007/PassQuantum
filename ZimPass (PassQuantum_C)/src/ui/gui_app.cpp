#include "zimpass/gui_app.hpp"

#include <imgui.h>
#include <imgui_impl_sdl2.h>
#include <imgui_impl_sdlrenderer2.h>
#include <SDL.h>

#include <array>
#include <stdexcept>

namespace zimpass {

GuiApp::GuiApp(Application& app)
    : app_(app) {}

int GuiApp::run() {
    if (SDL_Init(SDL_INIT_VIDEO | SDL_INIT_TIMER | SDL_INIT_GAMECONTROLLER) != 0) {
        throw std::runtime_error("SDL2 initialization failed");
    }

    SDL_Window* window = SDL_CreateWindow("ZimPass", SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, 1200, 760, SDL_WINDOW_RESIZABLE);
    SDL_Renderer* renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_ACCELERATED | SDL_RENDERER_PRESENTVSYNC);

    IMGUI_CHECKVERSION();
    ImGui::CreateContext();
    ImGuiIO& io = ImGui::GetIO();
    io.ConfigFlags |= ImGuiConfigFlags_NavEnableKeyboard;

    ImGui::StyleColorsDark();

    ImGui_ImplSDL2_InitForSDLRenderer(window, renderer);
    ImGui_ImplSDLRenderer2_Init(renderer);

    bool running = true;
    while (running) {
        SDL_Event event;
        while (SDL_PollEvent(&event)) {
            ImGui_ImplSDL2_ProcessEvent(&event);
            if (event.type == SDL_QUIT) {
                running = false;
            }
        }

        ImGui_ImplSDLRenderer2_NewFrame();
        ImGui_ImplSDL2_NewFrame();
        ImGui::NewFrame();

        draw_login_panel();
        draw_vault_panel();
        draw_settings_panel();

        ImGui::Render();
        SDL_SetRenderDrawColor(renderer, 24, 24, 26, 255);
        SDL_RenderClear(renderer);
        ImGui_ImplSDLRenderer2_RenderDrawData(ImGui::GetDrawData(), renderer);
        SDL_RenderPresent(renderer);
    }

    ImGui_ImplSDLRenderer2_Shutdown();
    ImGui_ImplSDL2_Shutdown();
    ImGui::DestroyContext();

    SDL_DestroyRenderer(renderer);
    SDL_DestroyWindow(window);
    SDL_Quit();

    return 0;
}

void GuiApp::draw_login_panel() {
    ImGui::Begin("Login");

    static char master_password[256] = {};
    ImGui::InputText("Master password", master_password, sizeof(master_password), ImGuiInputTextFlags_Password);

    if (ImGui::Button("Create profile")) {
        std::vector<std::uint8_t> placeholder_private_key(32, 0xAB);
        app_.create_master_profile(master_password, placeholder_private_key);
    }

    ImGui::SameLine();
    if (ImGui::Button("Unlock")) {
        app_.verify_master_password(master_password);
    }

    ImGui::Text("Unlocked: %s", app_.session().unlocked ? "yes" : "no");
    ImGui::End();
}

void GuiApp::draw_vault_panel() {
    ImGui::Begin("Vault");

    static char vault_path[512] = "vault.db";
    static char vault_key[256] = "change-me";

    ImGui::InputText("Vault path", vault_path, sizeof(vault_path));
    ImGui::InputText("Vault key", vault_key, sizeof(vault_key), ImGuiInputTextFlags_Password);

    if (ImGui::Button("Open SQLCipher vault")) {
        app_.open_vault(vault_path, vault_key);
    }

    if (ImGui::Button("Add sample entry")) {
        PasswordEntry e;
        e.id = 1;
        e.service = "example.com";
        e.username = "user@example.com";
        e.kyber_ciphertext = {1, 2, 3};
        e.nonce = {4, 5, 6};
        e.ciphertext = {7, 8, 9};
        app_.vault().upsert_password(e);
    }

    if (ImGui::Button("List entries")) {
        const auto entries = app_.vault().list_passwords();
        for (const auto& entry : entries) {
            ImGui::BulletText("%llu - %s (%s)", static_cast<unsigned long long>(entry.id), entry.service.c_str(), entry.username.c_str());
        }
    }

    ImGui::End();
}

void GuiApp::draw_settings_panel() {
    ImGui::Begin("Theme");
    static std::array<float, 4> accent {0.20f, 0.72f, 0.55f, 1.0f};
    ImGui::ColorEdit4("Accent", accent.data());
    ImGui::Text("Color customization pipeline wired through Dear ImGui.");
    ImGui::End();
}

} // namespace zimpass
