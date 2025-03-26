package main

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//go:embed default-pendulum.png
//go:embed gopher.png
var embedFS embed.FS

type pendulumData struct {
	l       float64
	m       float64
	k       float64
	g       float64
	angle   float64
	ang_spd float64
	t       float64
	dt      float64
}

type displays struct {
	time_display    *widget.Label
	angle_display   *widget.Label
	ang_spd_display *widget.Label
}

type PendulumLog struct {
	time_points    []float64
	angle_points   []float64
	ang_spd_points []float64
}

type pivotData struct {
	len float64
	x   float64
	y   float64
}

type spriteData struct {
	default_sprite *canvas.Image
	gopher_sprite  *canvas.Image
	gopher_mode    bool
}

func dataConvert(str string, value *float64) bool {
	var err error
	*value, err = strconv.ParseFloat(str, 64)
	if err == nil {
		return true
	} else {
		return false
	}
}

func parseInput(lInputField, mInputField, gInputField, kInputField, angleInputField, angSpdInputField *widget.Entry, data *pendulumData) bool {
	if !dataConvert(lInputField.Text, &data.l) || !dataConvert(mInputField.Text, &data.m) || !dataConvert(gInputField.Text, &data.g) || !dataConvert(kInputField.Text, &data.k) || !dataConvert(angleInputField.Text, &data.angle) || !dataConvert(angSpdInputField.Text, &data.ang_spd) || data.l == 0 || data.m == 0 {
		return false
	}
	return true
}

func iteration(data pendulumData, angle_old, pivot_len, pivot_x, pivot_y float64) (float64, float64, float64, float64, float64) {
	var accel, angle_new, x, y float64

	data.t += data.dt
	accel = data.g/data.l*math.Sin(data.angle) - data.k/data.m*(angle_old-data.angle)/data.dt
	angle_new = 2*data.angle - angle_old - accel*data.dt*data.dt
	angle_old = data.angle
	data.angle = angle_new
	x = float64(pivot_x) + pivot_len*math.Sin(data.angle)
	y = float64(pivot_y) + pivot_len*math.Cos(data.angle)
	return x, y, data.t, data.angle, angle_old
}

func animation(stopAnimation chan bool, plot_data PlotData, disp *displays, sprite_data *spriteData, line *canvas.Line, data pendulumData, pivot pivotData, pendulum_log *PendulumLog, running *bool) {
	var x, y float64
	gopher_mode_old := sprite_data.gopher_mode
	var current_sprite *canvas.Image
	if sprite_data.gopher_mode {
		current_sprite = sprite_data.gopher_sprite
	} else {
		current_sprite = sprite_data.default_sprite
	}

	angle_old := data.angle - data.ang_spd*data.dt

	go func() {
		for {
			if gopher_mode_old != sprite_data.gopher_mode {
				if sprite_data.gopher_mode {
					current_sprite = sprite_data.gopher_sprite
				} else {
					current_sprite = sprite_data.default_sprite
				}
				gopher_mode_old = sprite_data.gopher_mode
			}

			x, y, data.t, data.angle, angle_old = iteration(data, angle_old, pivot.len, pivot.x, pivot.y)

			disp.time_display.Text = fmt.Sprintf("Время: %.2f с", data.t)
			disp.time_display.Refresh()
			disp_angle := data.angle/math.Pi*180 - float64(360*(int(data.angle/math.Pi*180)/360))
			if disp_angle > 180 {
				disp_angle -= 360
			}
			disp.angle_display.Text = fmt.Sprintf("Угол: %.2f°", disp_angle)
			disp.angle_display.Refresh()
			disp.ang_spd_display.Text = fmt.Sprintf("Угловая скорость: %2.f град/с", (data.angle-angle_old)/data.dt/math.Pi*180)
			disp.ang_spd_display.Refresh()
			pendulum_log.time_points = append(pendulum_log.time_points, data.t)
			pendulum_log.angle_points = append(pendulum_log.angle_points, data.angle)
			pendulum_log.ang_spd_points = append(pendulum_log.ang_spd_points, (data.angle-angle_old)/data.dt)
			current_sprite.Move(fyne.NewPos(float32(x-25), float32(y-25)))
			line.Position2 = fyne.NewPos(float32(x), float32(y))
			if data.angle < 1e-6 && data.angle > -1e-6 && (data.angle-angle_old)*data.dt < 1e-6 && (data.angle-angle_old)*data.dt > -1e-6 {
				UpdatePlotTabs(plot_data, *pendulum_log)
				*running = false
				return
			}

			select {
			case <-stopAnimation:
				*running = false
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

func main() {
	a := app.New()
	w := a.NewWindow("Gopher pendulum")

	default_pendulum_img_data, _ := embedFS.ReadFile("default-pendulum.png")
	default_pendulum_img, _ := png.Decode(bytes.NewReader(default_pendulum_img_data))
	gopher_data, _ := embedFS.ReadFile("gopher.png")
	gopher_img, _ := png.Decode(bytes.NewReader(gopher_data))
	gopher_icon := fyne.NewStaticResource("gopher_icon", gopher_data)

	w.SetIcon(gopher_icon)

	stopAnimation := make(chan bool)

	pivot := pivotData{len: 170, x: 395, y: 195}

	data := pendulumData{dt: 0.01}

	var running bool
	var pendulum_log PendulumLog

	osc_graph_tab_content := container.NewWithoutLayout()
	osc_graph_tab := container.NewTabItem("График колебаний", osc_graph_tab_content)

	ang_spd_graph_tab_content := container.NewWithoutLayout()
	ang_spd_graph_tab := container.NewTabItem("График угловой скорости", ang_spd_graph_tab_content)

	phase_diagram_tab_content := container.NewWithoutLayout()
	phase_diagram_tab := container.NewTabItem("Фазовая диаграмма", phase_diagram_tab_content)

	plot_data := PlotData{osc_graph_tab_content, ang_spd_graph_tab_content, phase_diagram_tab_content}

	default_pendulum_sprite := canvas.NewImageFromImage(default_pendulum_img)
	default_pendulum_sprite.FillMode = canvas.ImageFillContain
	default_pendulum_sprite.Resize(fyne.NewSize(50, 50))
	default_pendulum_sprite.Move(fyne.NewPos(370, 340))

	gopher := canvas.NewImageFromImage(gopher_img)
	gopher.FillMode = canvas.ImageFillContain
	gopher.Resize(fyne.NewSize(0, 0))
	gopher.Move(fyne.NewPos(370, 340))

	rectangle := canvas.NewRectangle(color.White)
	rectangle.Resize(fyne.NewSize(800, 400))
	rectangle.Move(fyne.NewPos(-4, 0))

	pivot_sprite := canvas.NewCircle(color.Black)
	pivot_sprite.Resize(fyne.NewSize(11, 11))
	pivot_sprite.Move(fyne.NewPos(float32(pivot.x-5), float32(pivot.y-5)))

	circle := canvas.NewCircle(color.RGBA{0, 0, 150, 255})
	circle.Resize(fyne.NewSize(50, 50))
	circle.Move(fyne.NewPos(370, 340))

	line := canvas.NewLine(color.RGBA{111, 112, 111, 255})
	line.Position1 = fyne.NewPos(float32(pivot.x), float32(pivot.y))
	line.Position2 = fyne.NewPos(395, 365)
	line.StrokeWidth = 5

	l_label := widget.NewLabel("Длина, м")
	l_label.Move(fyne.NewPos(0, 400))

	lInputField := widget.NewEntry()
	lInputField.Resize(fyne.NewSize(100, 36))
	lInputField.Move(fyne.NewPos(5, 430))
	lInputField.Text = "1"

	m_label := widget.NewLabel("Масса, кг")
	m_label.Move(fyne.NewPos(105, 400))

	mInputField := widget.NewEntry()
	mInputField.Resize(fyne.NewSize(100, 36))
	mInputField.Move(fyne.NewPos(110, 430))
	mInputField.Text = "1"

	g_label := widget.NewLabel("g, м/с^2")
	g_label.Move(fyne.NewPos(210, 400))

	gInputField := widget.NewEntry()
	gInputField.Resize(fyne.NewSize(100, 36))
	gInputField.Move(fyne.NewPos(215, 430))
	gInputField.Text = "9.81"

	k_label := widget.NewLabel("Коэф сопр")
	k_label.Move(fyne.NewPos(315, 400))

	kInputField := widget.NewEntry()
	kInputField.Resize(fyne.NewSize(100, 36))
	kInputField.Move(fyne.NewPos(320, 430))
	kInputField.Text = "0.5"

	angle_label := widget.NewLabel("Начльн угол, °")
	angle_label.Move(fyne.NewPos(420, 400))

	angleInputField := widget.NewEntry()
	angleInputField.Resize(fyne.NewSize(100, 36))
	angleInputField.Move(fyne.NewPos(425, 430))
	angleInputField.Text = "70"

	ang_spd_label := widget.NewLabel("Начальн угл скор, град/с")
	ang_spd_label.Move(fyne.NewPos(525, 400))

	angSpdInputField := widget.NewEntry()
	angSpdInputField.Resize(fyne.NewSize(100, 36))
	angSpdInputField.Move(fyne.NewPos(530, 430))
	angSpdInputField.Text = "0"

	time_display := widget.NewLabel("Время: 0.00 c")
	time_display.Move(fyne.NewPos(110, 495))
	angle_display := widget.NewLabel("Угол: 0.00°")
	angle_display.Move(fyne.NewPos(110, 465))
	ang_spd_display := widget.NewLabel("Угловая скорость: 0.00 град/c")
	ang_spd_display.Move(fyne.NewPos(110, 480))

	disp := displays{time_display, angle_display, ang_spd_display}
	sprite_data := spriteData{default_sprite: default_pendulum_sprite, gopher_sprite: gopher}

	gopher_mode_chekbox := widget.NewCheck("Gopher mode", func(value bool) {
		sprite_data.gopher_mode = value
		if sprite_data.gopher_mode {
			sprite_data.gopher_sprite.Move(sprite_data.default_sprite.Position())
			sprite_data.gopher_sprite.Resize(fyne.NewSize(50, 50))
			sprite_data.default_sprite.Resize(fyne.NewSize(0, 0))
		} else {
			sprite_data.default_sprite.Move(sprite_data.gopher_sprite.Position())
			sprite_data.gopher_sprite.Resize(fyne.NewSize(0, 0))
			sprite_data.default_sprite.Resize(fyne.NewSize(50, 50))
		}
	})
	gopher_mode_chekbox.Resize(fyne.NewSize(36, 36))
	gopher_mode_chekbox.Move(fyne.NewPos(525, 465))

	startButton := widget.NewButton("Start", func() {
		if !running && parseInput(lInputField, mInputField, gInputField, kInputField, angleInputField, angSpdInputField, &data) {
			data.angle *= (3.14159 / 180)
			pendulum_log = PendulumLog{[]float64{}, []float64{}, []float64{}}
			if sprite_data.gopher_mode {
				animation(stopAnimation, plot_data, &disp, &sprite_data, line, data, pivot, &pendulum_log, &running)
			} else {
				animation(stopAnimation, plot_data, &disp, &sprite_data, line, data, pivot, &pendulum_log, &running)
			}
			running = true
		} else if running && parseInput(lInputField, mInputField, gInputField, kInputField, angleInputField, angSpdInputField, &data) {
			stopAnimation <- true
			pendulum_log = PendulumLog{[]float64{}, []float64{}, []float64{}}
			data.angle *= (3.14159 / 180)
			running = true
			if sprite_data.gopher_mode {
				animation(stopAnimation, plot_data, &disp, &sprite_data, line, data, pivot, &pendulum_log, &running)
			} else {
				animation(stopAnimation, plot_data, &disp, &sprite_data, line, data, pivot, &pendulum_log, &running)
			}
		}
	})
	startButton.Resize(fyne.NewSize(50, 30))
	startButton.Move(fyne.NewPos(5, 470))

	stopButton := widget.NewButton("Stop", func() {
		if running {
			stopAnimation <- true
			UpdatePlotTabs(plot_data, pendulum_log)
			pendulum_log = PendulumLog{}
		}
	})
	stopButton.Resize(fyne.NewSize(50, 30))
	stopButton.Move(fyne.NewPos(59, 470))

	closeButton := widget.NewButton("Quit", func() {
		a.Quit()
	})
	closeButton.Resize(fyne.NewSize(50, 30))
	closeButton.Move(fyne.NewPos(5, 504))

	main_tab_content := container.NewWithoutLayout(
		rectangle,
		line,
		pivot_sprite,
		sprite_data.default_sprite,
		sprite_data.gopher_sprite,
		l_label,
		lInputField,
		m_label,
		mInputField,
		g_label,
		gInputField,
		k_label,
		kInputField,
		angle_label,
		angleInputField,
		ang_spd_label,
		angSpdInputField,
		startButton,
		stopButton,
		closeButton,
		angle_display,
		ang_spd_display,
		time_display,
		gopher_mode_chekbox,
	)

	main_tab := container.NewTabItem("Анимация", main_tab_content)

	tabs := container.NewAppTabs(
		main_tab,
		osc_graph_tab,
		ang_spd_graph_tab,
		phase_diagram_tab,
	)

	tabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(tabs)

	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)

	w.ShowAndRun()
}
