package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type PlotData struct {
	osc_graph_tab_content     *fyne.Container
	ang_spd_graph_tab_content *fyne.Container
	phase_diagram_tab_content *fyne.Container
}

func UpdatePlotTabs(plot_data PlotData, pendulum_log PendulumLog) {
	drawGraph(plot_data.osc_graph_tab_content, pendulum_log.time_points, pendulum_log.angle_points, "α(t)", "t", "α", "angle_graph.png", true)
	drawGraph(plot_data.ang_spd_graph_tab_content, pendulum_log.time_points, pendulum_log.ang_spd_points, "ω(t)", "t", "ω", "ang_spd_graph.png", true)
	drawGraph(plot_data.phase_diagram_tab_content, pendulum_log.angle_points, pendulum_log.ang_spd_points, "Фазовая диаграмма", "α", "ω", "phase_diagram.png", true)
}

func drawGraph(tab_content *fyne.Container, x, y []float64, label, label_x, label_y, filename string, save bool) {
	plotPng(x, y, label, label_x, label_y, filename)
	graph := canvas.NewImageFromFile(filename)
	go func() {
		graph.Resize(fyne.NewSize(getImgDimansions(filename)))
		tab_content.Objects = []fyne.CanvasObject{graph}
		tab_content.Refresh()
	}()
	if !save {
		os.Remove(filename)
	}
}

func getImgDimansions(filename string) (float32, float32) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		panic(err)
	}
	return float32(config.Width), float32(config.Height)
}

func plotPng(x, y []float64, label, label_x, label_y, filename string) {
	p := plot.New()

	p.Title.Text = label
	p.X.Label.Text = label_x
	p.Y.Label.Text = label_y

	pts := make(plotter.XYs, len(x))
	for i := 0; i < len(x); i++ {
		pts[i].X = x[i]
		pts[i].Y = y[i]
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		log.Panic(err)
	}
	line.Color = color.RGBA{R: 255, A: 255}
	p.Add(line)

	err = p.Save(210*vg.Millimeter, 147*vg.Millimeter, filename)
	if err != nil {
		log.Panic(err)
	}
}
