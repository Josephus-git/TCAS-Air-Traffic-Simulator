# TCAS Air Traffic Simulator ✈️

## Overview

The TCAS Air Traffic Simulator is a graphical application designed to visualize air traffic flow and demonstrate the functionality of a Traffic Collision Avoidance System (TCAS). Users can configure the number of planes and simulation parameters to observe how TCAS prevents mid-air collisions, or how the absence/faulty TCAS leads to crashes.

## Features ✨

* **Configurable Simulations:** Set the number of planes and the duration of the simulation.
* **Altitude Control:** Choose between planes flying at a single cruise altitude or across three distinct altitudes (10,000 ft, 11,000 ft, 12,000 ft).
* **Real-time Rendering:** Dynamic visualization of airports icao:airport: and airplanes ✈️ in flight.
* **TCAS Warnings & Engagements:** Observe real-time visual indicators for TCAS warnings ⚠️ (close proximity) and engagements ✅ (imminent collision avoidance maneuvers).
* **Interactive Viewport:** Pan ↔️ and zoom 🔍 functionalities allow users to navigate and inspect specific areas of the simulation.
* **Command-Line Interface (CLI):** Interact with the simulation and control its state via a simple command-line interface.
* **Detailed Logging:** Comprehensive logs 📄 provide additional insights into simulation events and plane behaviors.

## Technologies Used 🛠️

* **Go:** The primary programming language for the simulation logic and application backend.
* **Fyne:** A cross-platform GUI toolkit for Go, used for rendering the graphical simulation and user interface.

## Installation and Running 🚀

To run the TCAS Air Traffic Simulator, ensure you have Go installed on your system. The Fyne toolkit and other Go dependencies should be automatically handled by Go Modules.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/Josephus-git/TCAS-Air-Traffic-Simulator.git
    cd tcas-air-traffic-simulator # Or whatever your project directory is named
    ```
2.  **Run the application:**
    ```bash
    go run .
    ```
    The `go run .` command will automatically download any required Fyne dependencies (as specified in `go.mod` and `go.sum`) if they are not already present.

## Usage 🎮

Upon launching the application, an input window will appear, allowing you to configure simulation parameters such as the number of planes and simulation duration.

Once configured, click "Start Simulation" to launch the graphical simulation window. You can interact with the simulation using the GUI controls (zoom, pan, home, quit) or via the command-line interface in your terminal.

### CLI Commands (Example)

After starting the simulation, you can interact with it via the terminal where you ran `go run .`:


TCAS-simulator > help

*(Further CLI commands will be listed here, e.g., `run`, `log all`, `exit`, etc.)*

## Screenshots & Videos 📸🎬

*(Add your screenshots and video demonstrations here to showcase the application in action. You can embed images directly or link to video hosting platforms.)*

---

## Contributing to TCAS Air Traffic Simulator 🤝

We welcome contributions to the TCAS Air Traffic Simulator! Whether it's reporting bugs, suggesting new features, or submitting code changes, your help is greatly appreciated.

### 🐛 Bug Reports

If you find a bug, please open an issue on our [Issues page]([Your Repository Issues URL Here]). When reporting a bug, please include:

* A clear and concise description of the bug.
* Steps to reproduce the behavior.
* Expected behavior.
* Actual behavior.
* Screenshots or videos if applicable.
* Your operating system and Go version.

### ✨ Feature Requests

Have an idea for a new feature or improvement? We'd love to hear it! Please open an issue on our [Issues page]([Your Repository Issues URL Here]) and describe your idea in detail.

## Contact ✉️

Feel free to connect with me:

* **LinkedIn:** [www.linkedin.com/in/josephus-otejere-2480282b5](https://www.linkedin.com/in/josephus-otejere-2480282b5)

## License 📜

This project is open-source and licensed under the **MIT License**. See the [LICENSE](LICENSE) file at the root of this repository for full details.
