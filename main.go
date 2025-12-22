package main

import (
	"math"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

// Символы
const (
	PlayerChar = '▲'
	EnemyChar  = 'M'
	BulletChar = '•'
	WallChar   = '▓' // Кирпичная стена
)

// Размеры и константы
const (
	Width  = 40
	Height = 20
)

// Настройки скорости движения (увеличьте, чтобы замедлить движение)
const (
	PlayerMoveDelay = 2 // number of ticks between player moves (70ms * n)
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Point struct {
	X, Y int
}

type Object struct {
	Pos Point
	Dir Direction
	id  int // Для идентификации врагов
}

type Bullet struct {
	Pos    Point
	Dir    Direction
	Active bool
}

// Глобальное состояние
var (
	player             Object
	enemies            []*Object // Используем указатели для удобства изменения
	bullets            []*Bullet
	walls              map[Point]bool // Карта стен для быстрого поиска
	score              int
	isOver             bool
	playerMoveCooldown int
	rng                *rand.Rand
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	initGame()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	ticker := time.NewTicker(70 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				if ev.Key == termbox.KeyEsc {
					return
				}
				if !isOver {
					handleInput(ev)
				} else {
					initGame()
				}
			}
		case <-ticker.C:
			if !isOver {
				updateState()
				draw()
			} else {
				drawGameOver()
			}
		}
	}
}

func initGame() {
	player = Object{Pos: Point{X: 2, Y: Height / 2}, Dir: Right}
	enemies = []*Object{}
	bullets = []*Bullet{}
	walls = make(map[Point]bool)
	score = 0
	isOver = false
	playerMoveCooldown = 0

	generateWalls()

	// Спавним 3 врагов
	for i := 0; i < 3; i++ {
		spawnEnemy()
	}
}

func generateWalls() {
	// Генерируем несколько случайных стен
	for i := 0; i < 40; i++ {
		x := rng.Intn(Width-2) + 1
		y := rng.Intn(Height-2) + 1
		// Не ставим стены на старте игрока
		if x < 5 {
			continue
		}
		walls[Point{X: x, Y: y}] = true
	}

	// Вертикальная стена посередине
	for y := 5; y < 15; y++ {
		walls[Point{X: Width / 2, Y: y}] = true
	}
}

func spawnEnemy() {
	// Враги появляются справа
	x := Width - 2
	y := rng.Intn(Height-2) + 1

	// Проверка, чтобы не заспавнить в стене
	if walls[Point{X: x, Y: y}] {
		y++ // Сдвигаем, если занято (упрощенно)
	}

	enemies = append(enemies, &Object{
		Pos: Point{X: x, Y: y},
		Dir: Left,
		id:  rng.Int(),
	})
}

func handleInput(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyArrowUp:
		player.Dir = Up
	case termbox.KeyArrowDown:
		player.Dir = Down
	case termbox.KeyArrowLeft:
		player.Dir = Left
	case termbox.KeyArrowRight:
		player.Dir = Right
	case termbox.KeySpace:
		fireBullet(player.Pos, player.Dir)
	}
}

// Попытка движения с проверкой границ и стен
func tryMove(obj *Object) bool {
	newPos := obj.Pos
	switch obj.Dir {
	case Up:
		newPos.Y--
	case Down:
		newPos.Y++
	case Left:
		newPos.X--
	case Right:
		newPos.X++
	}

	// 1. Проверка границ экрана
	if newPos.X <= 0 || newPos.X >= Width-1 || newPos.Y <= 0 || newPos.Y >= Height-1 {
		return false
	}

	// 2. Проверка стен
	if walls[newPos] {
		return false
	}

	obj.Pos = newPos
	return true
}

func fireBullet(pos Point, dir Direction) {
	b := &Bullet{Pos: pos, Dir: dir, Active: true}
	// Двигаем сразу, чтобы пуля появилась перед танком
	moveBullet(b)
	if b.Active { // Если сразу не врезалась в стену
		bullets = append(bullets, b)
	}
}

func moveBullet(b *Bullet) {
	switch b.Dir {
	case Up:
		b.Pos.Y--
	case Down:
		b.Pos.Y++
	case Left:
		b.Pos.X--
	case Right:
		b.Pos.X++
	}

	// Уход за границы
	if b.Pos.X <= 0 || b.Pos.X >= Width-1 || b.Pos.Y <= 0 || b.Pos.Y >= Height-1 {
		b.Active = false
		return
	}

	// Попадание в стену
	if walls[b.Pos] {
		b.Active = false
		delete(walls, b.Pos) // Разрушаем стену!
	}
}

// AI Врага
func updateEnemyAI(e *Object) {
	// 1. Простой AI: Пытаемся выровняться с игроком по одной из осей

	diffX := player.Pos.X - e.Pos.X
	diffY := player.Pos.Y - e.Pos.Y

	// С вероятностью 20% меняем тактику на случайную, чтобы не застревали
	if rng.Float32() < 0.2 {
		e.Dir = Direction(rng.Intn(4))
		// Пытаемся сдвинуться в случайном направлении, результат можно игнорировать
		tryMove(e)
	} else {
		// Логика преследования
		// Если по X мы далеко, пытаемся ехать по X
		if math.Abs(float64(diffX)) > math.Abs(float64(diffY)) {
			if diffX > 0 {
				e.Dir = Right
			} else {
				e.Dir = Left
			}
		} else {
			// Иначе по Y
			if diffY > 0 {
				e.Dir = Down
			} else {
				e.Dir = Up
			}
		}

		if !tryMove(e) {
			// Если уперлись (в стену), пробуем рандомное направление
			e.Dir = Direction(rng.Intn(4))
			tryMove(e)
		}
	}

	// 2. Стрельба
	// Если игрок на одной линии с врагом, враг стреляет с шансом 10%
	if (e.Pos.X == player.Pos.X || e.Pos.Y == player.Pos.Y) && rng.Float32() < 0.1 {
		fireBullet(e.Pos, e.Dir)
	}
}

func updateState() {
	// Пули
	activeBullets := []*Bullet{}
	for _, b := range bullets {
		moveBullet(b)
		if b.Active {
			activeBullets = append(activeBullets, b)
		}
	}
	bullets = activeBullets

	// Движение игрока: троттлинг по кадрам
	if playerMoveCooldown <= 0 {
		tryMove(&player)
		playerMoveCooldown = PlayerMoveDelay
	} else {
		playerMoveCooldown--
	}

	// Враги
	for _, e := range enemies {
		// Замедляем врагов (двигаются каждый 3-й кадр, чтобы игрок был быстрее)
		if rng.Float32() < 0.25 {
			updateEnemyAI(e)
		}

		if e.Pos == player.Pos {
			isOver = true
		}
	}

	// Коллизии пуль с танками
	activeEnemies := []*Object{}
	for _, e := range enemies {
		hit := false
		for _, b := range bullets {
			if b.Active && b.Pos == e.Pos {
				hit = true
				b.Active = false
				score += 100
				break
			}
		}
		if !hit {
			activeEnemies = append(activeEnemies, e)
		}
	}
	enemies = activeEnemies

	// Проверка: убила ли пуля игрока?
	for _, b := range bullets {
		if b.Active && b.Pos == player.Pos {
			isOver = true
		}
	}

	// Респавн
	if len(enemies) < 3 {
		if rng.Float32() < 0.05 { // Не мгновенный респавн
			spawnEnemy()
		}
	}
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Границы
	for x := 0; x < Width; x++ {
		termbox.SetCell(x, 0, '─', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(x, Height-1, '─', termbox.ColorWhite, termbox.ColorDefault)
	}
	for y := 0; y < Height; y++ {
		termbox.SetCell(0, y, '│', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(Width-1, y, '│', termbox.ColorWhite, termbox.ColorDefault)
	}

	// Стены
	for p := range walls {
		termbox.SetCell(p.X, p.Y, WallChar, termbox.ColorMagenta, termbox.ColorDefault)
	}

	// Игрок
	drawTank(player, termbox.ColorGreen)

	// Враги
	for _, e := range enemies {
		drawTank(*e, termbox.ColorRed)
	}

	// Пули
	for _, b := range bullets {
		termbox.SetCell(b.Pos.X, b.Pos.Y, BulletChar, termbox.ColorYellow, termbox.ColorDefault)
	}

	// Интерфейс
	drawText(1, Height, "ESC:Exit SPACE:Fire", termbox.ColorWhite)
	drawText(Width-10, Height, "Score:", termbox.ColorCyan)
	drawNumber(Width-3, Height, score, termbox.ColorCyan)

	termbox.Flush()
}

func drawTank(obj Object, color termbox.Attribute) {
	var ch rune
	switch obj.Dir {
	case Up:
		ch = '▲'
	case Down:
		ch = '▼'
	case Left:
		ch = '◄'
	case Right:
		ch = '►'
	}
	termbox.SetCell(obj.Pos.X, obj.Pos.Y, ch, color, termbox.ColorDefault)
}

func drawGameOver() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	msg := "GAME OVER"
	drawText(Width/2-len(msg)/2, Height/2, msg, termbox.ColorRed)
	drawText(Width/2-6, Height/2+1, "Score:", termbox.ColorWhite)
	drawNumber(Width/2+1, Height/2+1, score, termbox.ColorWhite)
	termbox.Flush()
}

func drawText(x, y int, text string, fg termbox.Attribute) {
	for i, c := range text {
		termbox.SetCell(x+i, y, c, fg, termbox.ColorDefault)
	}
}

func drawNumber(x, y, num int, fg termbox.Attribute) {
	s := []rune{}
	if num == 0 {
		s = append(s, '0')
	}
	for num > 0 {
		s = append([]rune{rune('0' + num%10)}, s...)
		num /= 10
	}
	for i, c := range s {
		termbox.SetCell(x+i, y, c, fg, termbox.ColorDefault)
	}
}
