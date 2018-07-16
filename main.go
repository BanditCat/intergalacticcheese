package main

import (
	"./engine"
	"./input"
	"fmt"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"math/rand"
	"time"
)

var vertexShader = `
#version 330
uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float slices;
uniform float curSlice;
in vec3 vert;
in float mass;
out float z;
out float scale;
out vec3 color;
void main() {
  float c = ( mass - 0.5 )/ 16.5;
  color = clamp( vec3( 1.0 - c, 0.2, c /2 + 0.5 ), 0.0, 1.0 );
  if( mass != 0.0 ){
    color *= c;
  }else{
    color = vec3(0);
  }
  scale = mod( gl_VertexID/2-1-curSlice, slices + 1) / (slices+1);
  z = length( camera * model * vec4( vert, 1 ) );
  gl_Position = projection * camera * model * vec4(vert, 1);
}
`

var fragmentShader = `
#version 330
out vec4 outputColor;
uniform float slices;
in vec3 color;
in float scale;
in float z;
void main() {
  vec3 ocolor = color * vec3(90/(z*z)) * vec3(scale);
  if( scale < ( 1.25 / slices ) )
    discard;
  else
    outputColor = vec4( ocolor, 1.0);
}
`

const numStars = 768
const numSlices = 150
const ticksPerSlice = 2
const subs = 16

const lineWidth = 8

const forceScale = 0.0005
const shipBrakes = 0.95
const mouseScale = (1 / 5000.0) / forceScale

type mainApp struct {
	rotx, roty    float32
	stars         []mgl32.Vec3
	oldStars      []mgl32.Vec3
	velocities    []mgl32.Vec3
	starMasses    []float32
	starArray     []float32
	starMassArray []float32
	counter       int
	paused        bool
	ship          BHShip
	lastShip      BHShip
	calcchan      chan int
}

type BHShip struct {
	orientation, rotation mgl32.Quat
	position, velocity    mgl32.Vec3
	// Desired reference frame
	dposition, dvelocity    mgl32.Vec3
	dorientation, drotation mgl32.Quat
	// Control mode
	mode                    int
}

func (this *BHShip) tick() {
	this.orientation = this.orientation.Mul(this.rotation)
	this.position = this.position.Add(this.velocity)
}
func (this *BHShip) force(rot mgl32.Quat, accel mgl32.Vec3) {
	this.rotation = this.rotation.Mul(rot)
	this.velocity = this.velocity.Add(accel)
}

func (this *mainApp) updateStarsSub(start, end int, ch chan int) {
	for i := start; i < end; i++ {
		for j := 0; j < numStars; j++ {
			if i != j {
				dif := this.oldStars[j].Sub(this.oldStars[i])
				dist := dif.Len() * 10
				dif = dif.Normalize()
				dif = dif.Mul((this.starMasses[j] * 0.00001) / (dist * dist))
				this.velocities[i] = this.velocities[i].Add(dif)
			}
		}
		this.stars[i] = this.oldStars[i].Add(this.velocities[i])
		if this.stars[i].Len() > 20 && this.velocities[i].Len() > 0.0001 {
			this.velocities[i] = mgl32.Vec3{}
		}
	}
	ch <- 0
}

func (this *mainApp) updateStars(ch chan int) {
	this.oldStars = this.stars
	this.stars = make([]mgl32.Vec3, numStars)
	start := 0
	for i := 0; i < subs; i++ {
		end := (i + 1) * numStars / subs
		go this.updateStarsSub(start, end, ch)
		start = end
	}
}
func (this *mainApp) genLines(engine *engine.Engine) {
	curSlice := this.counter / ticksPerSlice
	if this.counter%ticksPerSlice == 0 {
		if curSlice == numSlices {
			this.counter = 0
			curSlice = 0
		}
		engine.UniformFloat("main", "curSlice", float32(curSlice))
	}
	for i := 0; i < numStars; i++ {
		for j := 0; j < 3; j++ {
			index1 := ((curSlice*6 + j + 3) % (numSlices * 6)) + (i * 6 * (numSlices + 1))
			index2 := ((curSlice*6 + j + 6) % (numSlices * 6)) + (i * 6 * (numSlices + 1))
			this.starArray[index1] = this.oldStars[i][j]
			this.starArray[index2] = this.oldStars[i][j]
		}
		index1 := ((curSlice*2 + 1) % (numSlices * 2)) + (i * 2 * (numSlices + 1))
		index2 := ((curSlice*2 + 2) % (numSlices * 2)) + (i * 2 * (numSlices + 1))
		this.starMassArray[index1] = this.starMasses[i]
		this.starMassArray[index2] = this.starMasses[i]
	}
	this.counter++
}
func (this *mainApp) starInit() {
	this.starArray = make([]float32, numStars*6*(numSlices+1))
	this.starMassArray = make([]float32, numStars*2*(numSlices+1))
	this.starMasses = make([]float32, numStars)
	this.stars = make([]mgl32.Vec3, numStars)
	this.velocities = make([]mgl32.Vec3, numStars)
	for i := 0; i < numStars; i++ {
		this.stars[i] = mgl32.Vec3{rand.Float32() + 0.1, (rand.Float32()*2 - 1) * 0.2, 0.0}
		this.velocities[i] = mgl32.Vec3{0.0, 0.0, float32(math.Sqrt(0.00003 / float64(this.stars[i][0])))}
		angle := rand.Float32() * 2 * math.Pi

		this.stars[i] = (mgl32.Rotate3DY(angle)).Mul3x1(this.stars[i])
		this.velocities[i] = mgl32.Rotate3DY(angle).Mul3x1(this.velocities[i])
		this.starMasses[i] = float32(math.Pow(2, rand.Float64()*4))
	}
	this.oldStars = this.stars
}
func (this *mainApp) Init(engine *engine.Engine, input *input.Input) {
	fmt.Println("Init start!")
	this.ship.orientation = mgl32.QuatIdent()
	this.ship.rotation = mgl32.QuatIdent()

	gp := &input.GamePads[0]
	gp.Swap(&gp.RS, &gp.LB)
	gp.SwapDpad(mgl32.Vec2{0, 1}, &gp.A)
	gp.SwapDpad(mgl32.Vec2{0, -1}, &gp.Y)
	gp.SwapDpad(mgl32.Vec2{1, 0}, &gp.B)
	gp.SwapDpad(mgl32.Vec2{-1, 0}, &gp.X)

	last := glfw.GetTime()

	engine.MakeProgramOrPanic("main", vertexShader, fragmentShader)
	engine.UseProgram("main")
	mod := mgl32.Ident4()
	engine.UniformMatrix("main", "model", mod)
	engine.UniformFloat("main", "slices", float32(numSlices))
	engine.FragLocation("main", "outputColor")

	engine.GrabMouse(true)
	rand.Seed(time.Now().UnixNano())
	this.starInit()
	fmt.Printf("Init took %v", last-glfw.GetTime())
}
func (this *mainApp) Tick(engine *engine.Engine, input *input.Input, delta float32) bool {

	// Ship
	{
		var accel, aaccel mgl32.Vec3
		aaccel[2] = float32(0)
		aaccel[0] = -input.GamePads[0].LeftStick.X() + input.Mouse.Delta.X()*mouseScale
		aaccel[1] = input.GamePads[0].LeftStick.Y() + input.Mouse.Delta.Y()*mouseScale
		accel[0] = input.GamePads[0].RightStick.X()
		accel[1] = input.GamePads[0].RightStick.Y()
		accel[2] = input.GamePads[0].LeftTrigger - input.GamePads[0].RightTrigger
		if input.GamePads[0].LB || engine.GetKey(glfw.KeyQ) {
			aaccel[2] += 1
		}
		if input.GamePads[0].RB || engine.GetKey(glfw.KeyE) {
			aaccel[2] -= 1
		}
		if engine.GetKey(glfw.KeyD) {
			accel[0] += 1
		}
		if engine.GetKey(glfw.KeyA) {
			accel[0] -= 1
		}
		if engine.GetKey(glfw.KeyW) {
			accel[1] += 1
		}
		if engine.GetKey(glfw.KeyS) {
			accel[1] -= 1
		}
		if engine.GetKey(glfw.KeyLeftControl) {
			accel[2] += 1
		}
		if engine.GetKey(glfw.KeyLeftShift) {
			accel[2] -= 1
		}
		if accel.Len() > 1 {
			accel = accel.Normalize()
		}
		if accel.Len() > 1 {
			aaccel = aaccel.Normalize()
		}
		forcevec := mgl32.Vec3{accel[0] * forceScale, accel[1] * forceScale, accel[2] * forceScale}
		forcevec = this.ship.orientation.Mat4().Mul4x1(forcevec.Vec4(1)).Vec3()

		rot := mgl32.QuatRotate(aaccel[1]*forceScale, mgl32.Vec3{1, 0, 0})
		rot = rot.Mul(mgl32.QuatRotate(aaccel[0]*forceScale, mgl32.Vec3{0, 1, 0}))
		rot = rot.Mul(mgl32.QuatRotate(aaccel[2]*forceScale, mgl32.Vec3{0, 0, 1}))
		this.ship.force(rot, forcevec)
		this.ship.tick()
	}
	engine.UseProgram("main")
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	//this.rotx += input.GamePads[0].LeftStick.X() * 5 * delta
	//this.roty += input.GamePads[0].LeftStick.Y() * 5 * delta
	//this.rotx += input.Mouse.Delta[0]/500 + input.Mouse.Scroll[0]/10
	//this.roty += input.Mouse.Delta[1]/500 + input.Mouse.Scroll[1]/10
	modelX := mgl32.HomogRotate3D(float32(this.rotx), mgl32.Vec3{0, 1, 0})
	modelY := mgl32.HomogRotate3D(float32(this.roty), mgl32.Vec3{0, 0, 1})
	modelX = modelX.Mul4(modelY)
	engine.UniformMatrix("main", "model", modelX)
	proj := mgl32.Perspective(mgl32.DegToRad(45), engine.Height/engine.Width, 0.1, 100)
	engine.UniformMatrix("main", "projection", proj)
	camera := this.ship.orientation.Inverse().Mat4()
	camera = camera.Mul4(mgl32.Translate3D(-this.ship.position[0], -this.ship.position[1], -this.ship.position[2]))

	engine.UniformMatrix("main", "camera", camera)
	gl.LineWidth(lineWidth)

	quit := false
	{
		if !this.paused {
			if this.calcchan != nil {
				for i := 0; i < subs; i++ {
					<-this.calcchan
				}
			} else {
				fmt.Println("Make!")
				this.calcchan = make(chan int)
			}

			this.updateStars(this.calcchan)
			this.genLines(engine)
		}
		engine.SetBuffer("main", "vert", this.starArray, 3)
		engine.SetBuffer("main", "mass", this.starMassArray, 1)
	}
	gl.DrawArrays(gl.LINES, 0, int32(numStars*2*(numSlices+1)))
	if input.Mouse.Left && !engine.IsMouseGrabbed() {
		engine.GrabMouse(true)
	}
	if engine.GetKeyPressed(glfw.KeyR) || input.GamePads[0].AP {
		this.starInit()
	}
	if engine.GetKey(glfw.KeyF) || input.GamePads[0].Y {
		//this.ship.movement = mgl32.Ident4()
	}
	if engine.GetKeyPressed(glfw.KeySpace) || input.GamePads[0].BP {
		this.paused = !this.paused
	}
	if engine.GetKeyPressed(glfw.KeyEscape) {
		if engine.IsMouseGrabbed() {
			engine.GrabMouse(false)
		} else {
			quit = true
		}
	}
	if quit || input.GamePads[0].Select {
		return false
	} else {
		return true
	}
}
func (this *mainApp) Quit(engine *engine.Engine) {
	fmt.Println("Quit!")

}

func main() {
	fmt.Println("start!")
	var m mainApp
	engine := engine.Engine{App: &m, Width: 1024*4.0/3.0, Height: 1024, Title: "Intergallactic Cheese!!!"}

	for engine.Tick() {
		// lol time.Sleep(10000000)
	}
}
