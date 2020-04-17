package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	ScreenWidth  = 320
	ScreenHeight = 240
)

func CollideRect(x1, y1, w1, h1, x2, y2, w2, h2 int) (dx, dy int) {
	dx1, dx2 := x1-(x2+w2), x2-(x1+w1)
	if dx1 < 0 && dx2 < 0 {
		if dx1 > dx2 {
			dx = -dx1
		} else {
			dx = dx2
		}
	}
	dy1, dy2 := y1-(y2+h2), y2-(y1+h1)
	if dy1 < 0 && dy2 < 0 {
		if dy1 > dy2 {
			dy = -dy1
		} else {
			dy = dy2
		}
	}
	return
}

type CellType int

const (
	CellNone = iota
	CellWall
	CellSoft
	CellBomb
	CellFlame
)

func (ct CellType) Color() color.Color {
	switch ct {
	case CellNone:
		return color.Transparent
	case CellWall:
		return color.RGBA{0x30, 0x40, 0x50, 0xff}
	case CellSoft:
		return color.RGBA{0x40, 0x30, 0x20, 0xff}
	case CellBomb:
		return color.RGBA{0x20, 0x20, 0x20, 0xff}
	case CellFlame:
		return color.RGBA{0xf0, 0x20, 0x20, 0xff}
	default:
		return color.Transparent
	}
}

type Cell interface {
	Type() CellType
	Update(x, y int, m *Map)
	HasExpired() bool
}

type WallCell struct {
}

func NewWallCell() *WallCell {
	return &WallCell{}
}

func (c *WallCell) Type() CellType {
	return CellWall
}

func (c *WallCell) Update(x, y int, m *Map) {
}

func (c *WallCell) HasExpired() bool {
	return false
}

type SoftCell struct {
}

func NewSoftCell() *SoftCell {
	return &SoftCell{}
}

func (c *SoftCell) Type() CellType {
	return CellSoft
}

func (c *SoftCell) Update(x, y int, m *Map) {
}

func (c *SoftCell) HasExpired() bool {
	return false
}

type BombCell struct {
	framesToExplosion int
	power             int
	putter            *Character
}

func NewBombCell(c *Character) *BombCell {
	return &BombCell{
		framesToExplosion: 180,
		power:             3,
		putter:            c,
	}
}

func (c *BombCell) Type() CellType {
	return CellBomb
}

func (c *BombCell) Update(x, y int, m *Map) {
	c.framesToExplosion--
	if c.framesToExplosion == 0 {
		m.Explosion(c, x, y)
	}
}

func (c *BombCell) HasExpired() bool {
	return c.framesToExplosion <= 0
}

type FlameCell struct {
	frames int
}

func NewFlameCell() *FlameCell {
	return &FlameCell{frames: 60}
}

func (c *FlameCell) Type() CellType {
	return CellFlame
}

func (c *FlameCell) Update(x, y int, m *Map) {
	c.frames--
}

func (c *FlameCell) HasExpired() bool {
	return c.frames < 0
}

type Map struct {
	CellLength    int
	Width, Height int
	Cells         []Cell // Width * Height
}

func NewMap() *Map {
	length := 16
	m := &Map{CellLength: length, Width: ScreenWidth / length, Height: ScreenHeight / length}
	m.Cells = make([]Cell, m.Width*m.Height)

	// set block to border.
	for x := 0; x < m.Width; x++ {
		for y := 0; y < m.Height; y++ {
			*m.At(x, y) = NewSoftCell()
		}
	}

	for x := 0; x < m.Width; x++ {
		*m.At(x, 0) = NewWallCell()
		*m.At(x, m.Height-1) = NewWallCell()
	}
	for y := 0; y < m.Height; y++ {
		*m.At(0, y) = NewWallCell()
		*m.At(m.Width-1, y) = NewWallCell()
	}
	for x := 2; x < m.Width; x += 2 {
		for y := 2; y < m.Height; y += 2 {
			*m.At(x, y) = NewWallCell()
		}
	}
	*m.At(1, 1) = nil
	*m.At(2, 1) = nil
	*m.At(1, 2) = nil
	return m
}

func (m *Map) At(x, y int) *Cell {
	if x < 0 || m.Width <= x || y < 0 || m.Height <= y {
		return nil
	}
	return &m.Cells[x+y*m.Width]
}

func (m *Map) AtPos(x, y int) *Cell {
	cx, cy := x/m.CellLength, y/m.CellLength
	return m.At(cx, cy)
}

func (m *Map) ForEach(fn func(int, int, Cell)) {
	for x := 0; x < m.Width; x++ {
		for y := 0; y < m.Height; y++ {
			if cell := m.At(x, y); *cell != nil {
				fn(x, y, *cell)
			}
		}
	}
}

func (m *Map) Cleanup() {
	for x := 0; x < m.Width; x++ {
		for y := 0; y < m.Height; y++ {
			if cell := m.At(x, y); *cell != nil {
				if (*cell).HasExpired() {
					*cell = nil
				}
			}
		}
	}
}

func (m *Map) Explosion(c *BombCell, x, y int) {
	c.framesToExplosion = 0
	f := func(x, y int) bool {
		c := m.At(x, y)
		if (*c) != nil && (*c).Type() == CellWall {
			return false
		} else if (*c) != nil && (*c).Type() == CellSoft {
			*c = NewFlameCell()
			return false
		} else if (*c) != nil && (*c).Type() == CellBomb {
			bc := (*c).(*BombCell)
			if !bc.HasExpired() {
				m.Explosion(bc, x, y)
			}
		}
		*c = NewFlameCell()
		return true
	}
	for i := 0; i < c.power; i++ {
		if !f(x-i, y) {
			break
		}
	}
	for i := 0; i < c.power; i++ {
		if !f(x+i, y) {
			break
		}
	}
	for i := 0; i < c.power; i++ {
		if !f(x, y-i) {
			break
		}
	}
	for i := 0; i < c.power; i++ {
		if !f(x, y+i) {
			break
		}
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (m *Map) PositionCorrection(c *Character) {
	for i := 0; i < 5; i++ {
		correctX, correctY := 0, 0
		maxIntersection := 0
		m.ForEach(func(x, y int, cell Cell) {
			if cell.Type() != CellWall && cell.Type() != CellBomb && cell.Type() != CellSoft {
				return
			}
			cx, cy := x*m.CellLength, y*m.CellLength
			colX, colY := CollideRect(c.X, c.Y, c.Length, c.Length, cx, cy, m.CellLength, m.CellLength)
			collides := colX != 0 && colY != 0
			if cell.Type() == CellBomb && cell.(*BombCell).putter == c {
				if collides {
					// 置いた人が自分だったら当たったことにしない
					return
				}
				// 自分が置いた爆弾に自分が乗ってなかったら置いた人情報をはずしておく
				cell.(*BombCell).putter = nil
				return
			}
			if collides {
				if maxIntersection < absInt(colX)+absInt(colY) {
					maxIntersection = absInt(colX) + absInt(colY)
					if absInt(colX) < absInt(colY) {
						correctX, correctY = colX, 0
					} else {
						correctX, correctY = 0, colY
					}
				}
			}
		})
		c.X += correctX
		c.Y += correctY
		if maxIntersection == 0 {
			break
		}
	}
}

func (m *Map) Update(g *Game) {
	m.ForEach(func(x, y int, cell Cell) {
		cell.Update(x, y, m)
	})
	m.Cleanup()
}

func (m *Map) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, float64(m.CellLength*m.Width), float64(m.CellLength*m.Height), color.RGBA{0x80, 0x80, 0x80, 0xff})
	m.ForEach(func(x, y int, cell Cell) {
		cx, cy := x*m.CellLength, y*m.CellLength
		clr := cell.Type().Color()
		ebitenutil.DrawRect(screen, float64(cx+1), float64(cy+1), float64(m.CellLength-2), float64(m.CellLength-2), clr)
	})
}

type Character struct {
	X, Y   int
	VX, VY int
	Length int
}

func (c *Character) Update() {
	c.X += c.VX
	c.Y += c.VY
}

type Bomb struct {
	Character
	TimeToExplosion int
}

const (
	KEY_1 = iota
	KEY_2
	KEY_L
	KEY_R
	KEY_U
	KEY_D
	KEY_MAX
)

type Input struct {
	press [][]bool
}

func NewInput() *Input {
	press := make([][]bool, 2)
	for i := 0; i < 2; i++ {
		press[i] = make([]bool, KEY_MAX)
	}
	i := &Input{press: press}
	return i
}

func (i *Input) Update() {
	for k := 0; k < KEY_MAX; k++ {
		i.press[1][k] = i.press[0][k]
	}
	i.press[0][KEY_1] = ebiten.IsKeyPressed(ebiten.KeyZ)
	i.press[0][KEY_2] = ebiten.IsKeyPressed(ebiten.KeyX)
	i.press[0][KEY_L] = ebiten.IsKeyPressed(ebiten.KeyLeft)
	i.press[0][KEY_R] = ebiten.IsKeyPressed(ebiten.KeyRight)
	i.press[0][KEY_U] = ebiten.IsKeyPressed(ebiten.KeyUp)
	i.press[0][KEY_D] = ebiten.IsKeyPressed(ebiten.KeyDown)
}

func (i *Input) IsPressed(key int) bool {
	return i.press[0][key]
}

func (i *Input) IsJustPressed(key int) bool {
	return i.press[0][key] && !i.press[1][key]
}

type Player struct {
	Character
	Input *Input
}

func NewPlayer() *Player {
	p := &Player{
		Character: Character{
			X:      20,
			Y:      20,
			Length: 12,
		},
		Input: NewInput(),
	}
	return p
}

func (p *Player) Update(g *Game) {
	p.Input.Update()

	if p.Input.IsPressed(KEY_L) {
		p.VX = -1
	} else if p.Input.IsPressed(KEY_R) {
		p.VX = 1
	} else {
		p.VX = 0
	}
	if p.Input.IsPressed(KEY_U) {
		p.VY = -1
	} else if p.Input.IsPressed(KEY_D) {
		p.VY = 1
	} else {
		p.VY = 0
	}

	if p.Input.IsJustPressed(KEY_1) {
		g.PutBomb(p)
	}

	p.Character.Update()
}

func (p *Player) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, float64(p.X), float64(p.Y), float64(p.Length), float64(p.Length), color.RGBA{0xff, 0xff, 0xff, 0xff})
}

type Game struct {
	Player *Player
	Map    *Map
}

func (g *Game) PutBomb(p *Player) bool {
	if cell := g.Map.AtPos(g.Player.X+g.Player.Length/2, g.Player.Y+g.Player.Length/2); cell != nil && *cell == nil {
		*cell = NewBombCell(&p.Character)
		return true
	}
	return false
}

func (g *Game) Update(screen *ebiten.Image) error {
	g.Map.Update(g)
	g.Player.Update(g)
	g.Map.PositionCorrection(&g.Player.Character)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Map.Draw(screen)
	g.Player.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("fps:%3.f tps:%3.f", ebiten.CurrentFPS(), ebiten.CurrentTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("s")
	if err := ebiten.RunGame(&Game{Player: NewPlayer(), Map: NewMap()}); err != nil {
		log.Fatal(err)
	}
}
