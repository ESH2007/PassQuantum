#include "zimpass/application.hpp"
#include "zimpass/gui_app.hpp"

#include <exception>
#include <iostream>

int main() {
    try {
        zimpass::Application app;
        zimpass::GuiApp gui(app);
        return gui.run();
    } catch (const std::exception& ex) {
        std::cerr << "Fatal error: " << ex.what() << '\n';
        return 1;
    }
}
