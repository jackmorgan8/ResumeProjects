# GIF Dithering in Go

Given a GIF file, this program will dither each frame, reconstructing the same image with a 2 or 4 color palette

## Features
- Uses the Floyd-Steinberg dithering algorithm
- 6 selectable palettes
- Asynchronous frame dithering via GoRoutines

## Demo
![Ito](https://github.com/jackmorgan8/ResumeProjects/blob/main/GolangDither/Examples/ito.gif "Input GIF")

![Ito2](https://github.com/jackmorgan8/ResumeProjects/blob/main/GolangDither/Examples/ito_dithered_2.gif "Output (Palette 1)")

![Ito1](https://github.com/jackmorgan8/ResumeProjects/blob/main/GolangDither/Examples/ito_dithered_1.gif "Output (Palette 3)")

![RP7](https://github.com/jackmorgan8/ResumeProjects/blob/main/GolangDither/Examples/rp7.gif "Input GIF")

![RP7](https://github.com/jackmorgan8/ResumeProjects/blob/main/GolangDither/Examples/rp7_dithered_2.gif "Output (Palette 1)")

![RP7](https://github.com/jackmorgan8/ResumeProjects/blob/main/GolangDither/Examples/rp7_dithered_1.gif "Output (Palette 3)")
