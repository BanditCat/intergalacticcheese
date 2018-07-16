package input

/*
#cgo LDFLAGS: -L./ -lXinput9_1_0
#include <xinput.h>
#include <WinError.h>
#include <stdio.h>

typedef struct{
  XINPUT_STATE p1, p2, p3, p4;
  int v1, v2, v3, v4;
} inputState;

static inputState states;

inputState getXInput( void  ){
  XINPUT_STATE state;

  XINPUT_STATE* xs[] = { &states.p1, &states.p2, &states.p3, &states.p4 };
  int* v[] = { &states.v1, &states.v2, &states.v3, &states.v4 };
  for( int i = 0; i < 4; ++i ){
    memset( xs[ i ], 0, sizeof( XINPUT_STATE ) );
    if( XInputGetState( i, &state ) == ERROR_SUCCESS ){
      *v[ i ] = 1;
      memcpy( xs[ i ], &state, sizeof( XINPUT_STATE ) );
    }else
      *v[ i ] = 0;
  }

  return states;
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

type GamePad struct {
	Active                                              bool
	LeftTrigger, RightTrigger                           float32
	LeftStick, RightStick, Dpad                         mgl32.Vec2
	Start, Select, X, Y, A, B, LB, RB, LS, RS           bool
	StartP, SelectP, XP, YP, AP, BP, LBP, RBP, LSP, RSP bool
	UpP, DownP, LeftP, RightP                           bool
	swaps                                               [][2]uint16
	swapSticks                                          bool
	swapTriggers                                        bool
}

func (this *GamePad) ResetSwaps() {
	this.swaps = make([][2]uint16, 0)
	this.swapSticks = false
	this.swapTriggers = false
}
func (this *GamePad) SwapSticks(v bool) {
	this.swapSticks = v
}
func (this *GamePad) SwapTriggers(v bool) {
	this.swapTriggers = v
}
func (this *GamePad) Swap(b1 *bool, b2 *bool) {
	if this.swaps == nil {
		this.swaps = make([][2]uint16, 0)
	}

	args := [2]*bool{b1, b2}
	ans := [2]uint16{}
	for i := 0; i < 2; i++ {
		switch args[i] {
		case &this.Start:
			ans[i] = 0x0010
		case &this.Select:
			ans[i] = 0x0020
		case &this.LS:
			ans[i] = 0x0040
		case &this.RS:
			ans[i] = 0x0080
		case &this.LB:
			ans[i] = 0x0100
		case &this.RB:
			ans[i] = 0x0200
		case &this.A:
			ans[i] = 0x1000
		case &this.B:
			ans[i] = 0x2000
		case &this.X:
			ans[i] = 0x4000
		case &this.Y:
			ans[i] = 0x8000
		}
	}
	this.swaps = append(this.swaps, ans)
}
func (this *GamePad) SwapDpad(dir mgl32.Vec2, b2 *bool) {
	if this.swaps == nil {
		this.swaps = make([][2]uint16, 0)
	}

	ans := [2]uint16{}
	if dir[0] > 0.9 {
		ans[0] = 0x0008
	} else if dir[0] < -0.9 {
		ans[0] = 0x0004
	} else {
		if dir[1] > 0.9 {
			ans[0] = 0x0001
		} else {
			ans[0] = 0x0002
		}
	}
	switch b2 {
	case &this.Start:
		ans[1] = 0x0010
	case &this.Select:
		ans[1] = 0x0020
	case &this.LS:
		ans[1] = 0x0040
	case &this.RS:
		ans[1] = 0x0080
	case &this.LB:
		ans[1] = 0x0100
	case &this.RB:
		ans[1] = 0x0200
	case &this.A:
		ans[1] = 0x1000
	case &this.B:
		ans[1] = 0x2000
	case &this.X:
		ans[1] = 0x4000
	case &this.Y:
		ans[1] = 0x8000
	}

	this.swaps = append(this.swaps, ans)
}

type Mouse struct {
	Delta               mgl32.Vec2
	Left, Middle, Right bool
	Scroll              mgl32.Vec2
}

type Input struct {
	GamePads     []GamePad
	lastGamePads []GamePad
	Mouse        Mouse
}

func (this Input) String() string {
	var buf bytes.Buffer
	for i := 0; i < len(this.GamePads); i++ {
		if this.GamePads[i].Active {
			buf.WriteString(fmt.Sprintf("\n\nController %d\n", i))
			buf.WriteString(fmt.Sprintf("Left trigger: %5.2f               Right trigger: %5.2f\n",
				this.GamePads[i].LeftTrigger, this.GamePads[i].RightTrigger))
			buf.WriteString(fmt.Sprintf("Left stick:   %5.2fx%5.2f|%5.2f|  Right stick:   %5.2fx%5.2f|%5.2f|\n",
				this.GamePads[i].LeftStick.X(), this.GamePads[i].LeftStick.Y(), this.GamePads[i].LeftStick.Len(),
				this.GamePads[i].RightStick.X(), this.GamePads[i].RightStick.Y(), this.GamePads[i].RightStick.Len()))
			buf.WriteString(fmt.Sprintf("Start: %5t Select: %5t LB: %5t RB: %5t dpad: %2.fx%2.f\n",
				this.GamePads[i].Start, this.GamePads[i].Select, this.GamePads[i].LB, this.GamePads[i].RB,
				this.GamePads[i].Dpad.X(), this.GamePads[i].Dpad.Y()))
			buf.WriteString(fmt.Sprintf("A: %5t B: %5t X: %5t Y: %5t LS:%5t RS: %5t\n",
				this.GamePads[i].A, this.GamePads[i].B, this.GamePads[i].X, this.GamePads[i].Y,
				this.GamePads[i].LS, this.GamePads[i].RS))

		}
	}

	return buf.String()
}

// Get populates the inputs slices.
func (this *Input) Get() {
	var swapss [4][][2]uint16
	var swapsst, swapstr [4]bool
	if len(this.GamePads) == 4 {
		swapss = [4][][2]uint16{
			this.GamePads[0].swaps,
			this.GamePads[1].swaps,
			this.GamePads[2].swaps,
			this.GamePads[3].swaps,
		}
		swapstr = [4]bool{
			this.GamePads[0].swapTriggers,
			this.GamePads[1].swapTriggers,
			this.GamePads[2].swapTriggers,
			this.GamePads[3].swapTriggers,
		}
		swapsst = [4]bool{
			this.GamePads[0].swapSticks,
			this.GamePads[1].swapSticks,
			this.GamePads[2].swapSticks,
			this.GamePads[3].swapSticks,
		}
	}
	this.lastGamePads = this.GamePads
	if this.lastGamePads == nil {
		this.lastGamePads = make([]GamePad,4)
	}
	this.GamePads = nil
	cstates := C.getXInput()
	states := [4]C.XINPUT_STATE{cstates.p1, cstates.p2, cstates.p3, cstates.p4}
	valids := [4]bool{int(cstates.v1) != 0, int(cstates.v2) != 0,
		int(cstates.v3) != 0, int(cstates.v4) != 0}
	for i := 0; i < 4; i++ {
		if valids[i] {

			// Get buttons and do swaps
			buttons := uint16(states[i].Gamepad.wButtons)
			for j := 0; j < len(swapss[i]); j++ {

				b1 := swapss[i][j][0]
				b2 := swapss[i][j][1]
				t1 := b1&buttons != 0
				t2 := b2&buttons != 0
				if t1 {
					buttons &^= b2
					buttons ^= b2
				} else {
					buttons &^= b2
				}
				if t2 {
					buttons &^= b1
					buttons ^= b1
				} else {
					buttons &^= b1
				}
			}

			var dx, dy float32 = 0, 0
			if buttons&0x0004 != 0 {
				dx -= 1
			}
			if buttons&0x0008 != 0 {
				dx += 1
			}
			if buttons&0x0002 != 0 {
				dy -= 1
			}
			if buttons&0x0001 != 0 {
				dy += 1
			}
			trigs := [2]float32{
				float32(states[i].Gamepad.bLeftTrigger) / 255.0,
				float32(states[i].Gamepad.bRightTrigger) / 255.0,
			}
			sticks := [2]mgl32.Vec2{
				mgl32.Vec2{
					(float32(states[i].Gamepad.sThumbLX) + 0.5) / 32767.5,
					(float32(states[i].Gamepad.sThumbLY) + 0.5) / 32767.5,
				},
				mgl32.Vec2{
					(float32(states[i].Gamepad.sThumbRX) + 0.5) / 32767.5,
					(float32(states[i].Gamepad.sThumbRY) + 0.5) / 32767.5,
				},
			}
			if swapsst[i] {
				temp := sticks[0]
				sticks[0] = sticks[1]
				sticks[1] = temp
			}
			if swapstr[i] {
				temp := trigs[0]
				trigs[0] = trigs[1]
				trigs[1] = temp
			}

			for i := 0; i < 2; i++ {
				if sticks[i].Len() > 1 {
					sticks[i] = sticks[i].Normalize()
				}
				if sticks[i].Len() < 0.1 {
					sticks[i][0] = 0
					sticks[i][1] = 0
				} else {
					dist := sticks[i].Len()
					dist -= 0.1
					dist *= 10.0 / 9.0
					sticks[i] = sticks[i].Normalize()
					sticks[i][0] *= dist
					sticks[i][1] *= dist
				}
			}

			this.GamePads = append(this.GamePads,
				GamePad{
					Active:       true,
					LeftTrigger:  trigs[0],
					RightTrigger: trigs[1],
					LeftStick:    sticks[0],
					RightStick:   sticks[1],
					Dpad:         mgl32.Vec2{dx, dy},
					Start:        buttons&0x0010 != 0,
					Select:       buttons&0x0020 != 0,
					LB:           buttons&0x0100 != 0,
					RB:           buttons&0x0200 != 0,
					LS:           buttons&0x0040 != 0,
					RS:           buttons&0x0080 != 0,
					A:            buttons&0x1000 != 0,
					B:            buttons&0x2000 != 0,
					X:            buttons&0x4000 != 0,
					Y:            buttons&0x8000 != 0,
					swaps:        swapss[i],
					swapTriggers: swapstr[i],
					swapSticks:   swapsst[i],
				})
		} else {
			this.GamePads = append(this.GamePads, GamePad{
				swaps:        swapss[i],
				swapTriggers: swapstr[i],
				swapSticks:   swapsst[i],
			})
		}
		if !this.lastGamePads[i].A && this.GamePads[i].A {
			this.GamePads[i].AP = true
		}
		if !this.lastGamePads[i].B && this.GamePads[i].B {
			this.GamePads[i].BP = true
		}
		if !this.lastGamePads[i].X && this.GamePads[i].X {
			this.GamePads[i].XP = true
		}
		if !this.lastGamePads[i].Y && this.GamePads[i].Y {
			this.GamePads[i].YP = true
		}
		if !this.lastGamePads[i].LB && this.GamePads[i].LB {
			this.GamePads[i].LBP = true
		}
		if !this.lastGamePads[i].RB && this.GamePads[i].RB {
			this.GamePads[i].RBP = true
		}
		if !this.lastGamePads[i].LS && this.GamePads[i].LS {
			this.GamePads[i].LSP = true
		}
		if !this.lastGamePads[i].RS && this.GamePads[i].RS {
			this.GamePads[i].RSP = true
		}
		if !this.lastGamePads[i].Start && this.GamePads[i].Start {
			this.GamePads[i].StartP = true
		}
		if !this.lastGamePads[i].Select && this.GamePads[i].Select {
			this.GamePads[i].SelectP = true
		}
		if this.lastGamePads[i].Dpad[0] != 1 && this.GamePads[i].Dpad[0] == 1 {
			this.GamePads[i].RightP = true
		}
		if this.lastGamePads[i].Dpad[0] != -1 && this.GamePads[i].Dpad[0] == -1 {
			this.GamePads[i].LeftP = true
		}
		if this.lastGamePads[i].Dpad[1] != 1 && this.GamePads[i].Dpad[1] == 1 {
			this.GamePads[i].UpP = true
		}
		if this.lastGamePads[i].Dpad[1] != -1 && this.GamePads[i].Dpad[1] == -1 {
			this.GamePads[i].DownP = true
		}
	}
}
