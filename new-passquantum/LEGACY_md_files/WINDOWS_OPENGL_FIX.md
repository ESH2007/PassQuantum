# Windows OpenGL Error - Quick Fix Guide

If you see this error when running PassQuantum on Windows:

```
Fyne error: window creation error
Cause: APIUnavailable: WGL: The driver does not appear to support OpenGL
```

## Quick Solutions

### ✅ Solution 1: Update Your Graphics Drivers (RECOMMENDED)

This error occurs when your graphics drivers are outdated or missing.

1. **Identify your graphics card:**
   - Press `Windows Key + R`
   - Type `dxdiag` and press Enter
   - Click on the "Display" tab to see your graphics card manufacturer

2. **Download latest drivers:**
   - **NVIDIA:** https://www.nvidia.com/Download/index.aspx
   - **AMD:** https://www.amd.com/en/support
   - **Intel:** https://www.intel.com/content/www/us/en/download-center/home.html

3. **Install the drivers and restart your computer**

### ✅ Solution 2: Use Software Rendering (If updating drivers doesn't work)

If you can't update drivers (old computer, virtual machine, etc.):

1. Download Mesa3D from: https://fdossena.com/?p=mesa/index.frag
2. Extract the file `opengl32.dll`
3. Put `opengl32.dll` in the same folder as `PassQuantum.exe`
4. Run `PassQuantum.exe`

**Note:** This uses software rendering and may be slower, but will work.

### ✅ Solution 3: Check Windows Updates

Sometimes Windows Update includes graphics driver updates:

1. Open Settings → Update & Security → Windows Update
2. Click "Check for updates"
3. Install any available updates
4. Restart your computer

## Still Having Issues?

Make sure you have:
- Windows 10 or later
- At least 4GB RAM
- An OpenGL 2.0 compatible graphics card (or the Mesa3D software renderer)

## Contact Support

If none of these solutions work, please report the issue with:
- Your Windows version
- Your graphics card model
- The complete error message
