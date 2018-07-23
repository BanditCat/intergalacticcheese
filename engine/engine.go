package engine

import (
	"../input"
	"fmt"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"runtime"
	"strings"
)

type App interface {
	Init(*Engine, *input.Input)
	Quit(*Engine)
	Tick(*Engine, *input.Input, float32) bool
}

type Engine struct {
	Title         string
	App           App
	Width, Height float32

	vao        uint32
	programs   map[string]uint32
	uniforms   map[string](map[string]uint32)
	attribs    map[string](map[string]uint32)
	buffers    map[string]uint32
	win        *glfw.Window
	inited     bool
	input      input.Input
	lastTime   float64
	lastCursor mgl32.Vec2
	scroll     mgl32.Vec2
	keyPresses map[glfw.Key]bool
}

func (this *Engine) scrollCallback(win *glfw.Window, xoff, yoff float64) {
	if xoff > 0 {
		this.scroll[0] += 1
	}
	if xoff < 0 {
		this.scroll[0] -= 1
	}
	if yoff > 0 {
		this.scroll[1] += 1
	}
	if yoff < 0 {
		this.scroll[1] -= 1
	}
}
func (this *Engine) init() {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		panic(err)
	}
		last := glfw.GetTime()
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.ContextCreationAPI, glfw.NativeContextAPI)

	if window, err := glfw.CreateWindow(int(this.Width), int(this.Height), this.Title, nil, nil); err != nil {
		panic(err)
	} else {
		this.win = window
	}
	
	this.win.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.GenVertexArrays(1, &(this.vao))
	gl.BindVertexArray(this.vao)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	this.programs = make(map[string]uint32)
	this.uniforms = make(map[string](map[string]uint32))
	this.attribs = make(map[string](map[string]uint32))
	this.buffers = make(map[string]uint32)
	this.keyPresses = make(map[glfw.Key]bool)
	this.lastTime = glfw.GetTime()
	{
		x, y := this.win.GetCursorPos()
		this.lastCursor[0], this.lastCursor[1] = float32(x), float32(y)
	}
	this.win.SetScrollCallback(this.scrollCallback)
	this.input.Get()
	
	this.App.Init(this,&this.input)
	fmt.Printf("\nEngine init took %v\n", last-glfw.GetTime())
}
func (this *Engine) GrabMouse(grab bool) {
	if grab {
		this.win.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		this.win.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}
func (this *Engine) IsMouseGrabbed() bool {
	return this.win.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}
func (this *Engine) Tick() bool {
	runtime.LockOSThread()
	if !this.inited {
		this.inited = true
		this.init()
	}

	now := glfw.GetTime()
	elapsed := float32(now - this.lastTime)
	this.lastTime = now

	gl.BindVertexArray(this.vao)

	this.input.Get()
	//Do mouse input here.
	{
		x64, y64 := this.win.GetCursorPos()
		x, y := float32(x64), float32(y64)
		this.input.Mouse.Left = this.win.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press
		this.input.Mouse.Right = this.win.GetMouseButton(glfw.MouseButtonRight) == glfw.Press
		this.input.Mouse.Middle = this.win.GetMouseButton(glfw.MouseButtonMiddle) == glfw.Press
		this.input.Mouse.Delta = mgl32.Vec2{this.lastCursor[0] - x, this.lastCursor[1] - y}
		this.input.Mouse.Scroll = this.scroll
		this.scroll = mgl32.Vec2{}
		this.lastCursor[0], this.lastCursor[1] = float32(x), float32(y)
	}

	h, w := this.win.GetSize()
	gl.Viewport(0, 0, int32(h), int32(w))
	this.Width = float32(w)
	this.Height = float32(h)

	if !this.App.Tick(this, &this.input, elapsed) || this.win.ShouldClose() {
		this.quit()
		return false
	}
	this.win.SwapBuffers()
	glfw.PollEvents()
	return true
}
func (this *Engine) quit() {
	this.App.Quit(this)
	glfw.Terminate()
}
func (this *Engine) GetKey(key glfw.Key) bool {
	return this.win.GetKey(key) == glfw.Press
}
func (this *Engine) GetKeyPressed(key glfw.Key) bool {
	ret := this.GetKey(key)
	if ret {
		if this.keyPresses[key] {
			return false
		} else {
			this.keyPresses[key] = true
			return true
		}
	} else {
		this.keyPresses[key] = false
		return false
	}
}
func (this *Engine) SetBuffer(prog, name string, data []float32, size int) {
	this.UseProgram(prog)
	buf, ok := this.buffers[name]
	if !ok {
		gl.GenBuffers(1, &buf)
		this.buffers[name] = buf
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, gl.Ptr(data), gl.STATIC_DRAW)
	attrib := this.getAttrib(prog, name)
	gl.EnableVertexAttribArray(attrib)
	gl.VertexAttribPointer(attrib, int32(size), gl.FLOAT, false, 0, gl.PtrOffset(0))
}
func (this *Engine) FragLocation(prog, out string) {
	this.UseProgram(prog)
	gl.BindFragDataLocation(this.programs[prog], 0, gl.Str(out+"\x00"))
}
func (this *Engine) UseProgram(prog string) {
	gl.UseProgram(this.programs[prog])
}
func (this *Engine) getLoc(program, uniform string) uint32 {
	this.UseProgram(program)
	if _, ok := this.uniforms[program]; !ok {
		this.uniforms[program] = make(map[string]uint32)
	}
	if _, ok := this.uniforms[program][uniform]; !ok {
		this.uniforms[program][uniform] =
			uint32(gl.GetUniformLocation(this.programs[program], gl.Str(uniform+"\x00")))
	}
	return this.uniforms[program][uniform]
}
func (this *Engine) getAttrib(program, attrib string) uint32 {
	this.UseProgram(program)
	if _, ok := this.attribs[program]; !ok {
		this.attribs[program] = make(map[string]uint32)
	}
	if _, ok := this.attribs[program][attrib]; !ok {
		this.attribs[program][attrib] =
			uint32(gl.GetAttribLocation(this.programs[program], gl.Str(attrib+"\x00")))
	}
	return this.attribs[program][attrib]
}

func (this *Engine) UniformMatrix(program, uniform string, matrix mgl32.Mat4) {
	uni := this.getLoc(program, uniform)
	gl.UniformMatrix4fv(int32(uni), 1, false, &matrix[0])
}
func (this *Engine) UniformFloat(program, uniform string, float float32) {
	uni := this.getLoc(program, uniform)
	gl.Uniform1f(int32(uni), float)
}

func (this *Engine) MakeProgramOrPanic(name, vertexShaderSource, fragmentShaderSource string) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteShader(vertexShader)

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}
	defer gl.DeleteShader(fragmentShader)

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		panic(fmt.Errorf("failed to link program: %v", log))
	}

	this.programs[name] = program
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
