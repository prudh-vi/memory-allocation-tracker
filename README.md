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


## How It Works

The application simulates memory allocation using two models:
1. **Paging**: Fixed-size memory blocks (4KB each)
2. **Segmentation**: Variable-size memory blocks

When you allocate memory (by pressing 'a'), the application:
- Finds contiguous free pages
- Allocates 1-3 pages at once
- Assigns a random process ID
- Updates the visualization

When you deallocate memory (by pressing 'd'), the application:
- Removes the last allocated segment
- Frees all pages associated with that process ID
- Updates the visualization

## Screenshots

(Add screenshots here)


## Requirements

- Go 1.16 or higher
- Required packages:
  - github.com/gdamore/tcell/v2
  - github.com/rivo/tview
  - github.com/shirou/gopsutil

## Installation

1. Ensure Go is installed on your system
2. Install the required dependencies:
