# go-mouse-mover

A tool that records mouse coordinates and automatically moves the mouse pointer to those coordinates in a loop. It runs in the system tray so you can easily see the current status and control the application from the taskbar.

## Usage

### Console

1. Enter `1` to start recording mouse coordinates.
2. Click the left mouse button at the desired positions. Multiple positions can be recorded.
3. Press `Ctrl+Shift+q` to stop recording.
4. Enter the interval in milliseconds between movements. For example, enter `1000` for a 1-second interval.
5. The recorded coordinates and interval are saved in `moveinfo.json`. This file can be reused later without re-recording.
6. Enter `2` to start automatic mouse movement.
7. Press `Ctrl+Shift+q` to stop the movement.

### System Tray

- The tray icon is **gray** when idle and **red** while moving.
- **Start** — Start automatic mouse movement using the saved `moveinfo.json`.
- **Stop** — Stop the ongoing automatic movement.
- **Quit** — Exit the application.

## License

go-mouse-mover is under the [MIT license](https://en.wikipedia.org/wiki/MIT_License)
