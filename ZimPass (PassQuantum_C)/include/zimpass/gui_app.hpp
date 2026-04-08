#pragma once

#include "zimpass/application.hpp"

namespace zimpass {

class GuiApp {
public:
    explicit GuiApp(Application& app);
    int run();

private:
    void draw_login_panel();
    void draw_vault_panel();
    void draw_settings_panel();

    Application& app_;
};

} // namespace zimpass
