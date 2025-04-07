# Memory Allocation Tracker

A terminal-based visualization tool for memory allocation and management with a Matrix-inspired interface.

## Overview

This application provides a real-time visualization of memory allocation concepts including:
- Paging
- Segmentation
- Memory usage statistics
- Fragmentation metrics

The interface is designed with a Matrix-inspired color scheme for a visually engaging experience.

## Features

- Real-time memory usage visualization
- Page allocation and deallocation simulation
- Memory segmentation visualization
- System memory statistics
- Fragmentation rate calculation
- Matrix-style animations

## Controls

- `a` - Allocate memory (random size)
- `d` - Deallocate memory (last allocated segment)
- `q` - Quit application
- `Esc` or `Ctrl+C` - Exit application

## Requirements

- Go 1.16 or higher
- Required packages:
  - github.com/gdamore/tcell/v2
  - github.com/rivo/tview
  - github.com/shirou/gopsutil

## Installation

1. Ensure Go is installed on your system
2. Install the required dependencies:
