// Copyright 2014 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Renders a textured spinning cube using GLFW 3 and OpenGL 4.1 core forward-compatible profile.
package main

import (
	"fmt"
	_ "image/png"
	"log"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const windowWidth = 1024
const windowHeight = 768

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Samples, 4)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Cube and Triangle", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow
	if err = gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	// Ensure we can capture the escape key being pressed below
	window.SetInputMode(glfw.StickyKeysMode, glfw.True)

	gl.ClearColor(0.0, 0.0, 0.4, 0)

	// Enable depth test
	gl.Enable(gl.DEPTH_TEST)
	// Accept fragment if it closer to the camera than the former one
	gl.DepthFunc(gl.LESS)

	// Configure the vertex data
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	viewUniform := gl.GetUniformLocation(program, gl.Str("view\x00"))
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(windowWidth)/windowHeight, 0.1, 10)
	view := mgl32.LookAt(4, 3, 3, 0, 0, 0, 0, 1, 0)
	model := mgl32.Ident4()
	model2 := mgl32.Translate3D(2, 0, 0)

	var cubeVerticies = []float32{
		//  X, Y, Z
		// Bottom
		-1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		-1.0, -1.0, 1.0,
		1.0, -1.0, -1.0,
		1.0, -1.0, 1.0,
		-1.0, -1.0, 1.0,

		// Top
		-1.0, 1.0, -1.0,
		-1.0, 1.0, 1.0,
		1.0, 1.0, -1.0,
		1.0, 1.0, -1.0,
		-1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,

		// Front
		-1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
		-1.0, 1.0, 1.0,

		// Back
		-1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0,
		1.0, 1.0, -1.0,

		// Left
		-1.0, -1.0, 1.0,
		-1.0, 1.0, -1.0,
		-1.0, -1.0, -1.0,
		-1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,

		// Right
		1.0, -1.0, 1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, -1.0,
		1.0, 1.0, 1.0,
	}

	var vboCube uint32
	gl.GenBuffers(1, &vboCube)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboCube)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeVerticies)*4, gl.Ptr(cubeVerticies), gl.STATIC_DRAW)

	var cubeColors = []float32{
		0.583, 0.771, 0.014,
		0.609, 0.115, 0.436,
		0.327, 0.483, 0.844,
		0.822, 0.569, 0.201,
		0.435, 0.602, 0.223,
		0.310, 0.747, 0.185,
		0.597, 0.770, 0.761,
		0.559, 0.436, 0.730,
		0.359, 0.583, 0.152,
		0.483, 0.596, 0.789,
		0.559, 0.861, 0.639,
		0.195, 0.548, 0.859,
		0.014, 0.184, 0.576,
		0.771, 0.328, 0.970,
		0.406, 0.615, 0.116,
		0.676, 0.977, 0.133,
		0.971, 0.572, 0.833,
		0.140, 0.616, 0.489,
		0.997, 0.513, 0.064,
		0.945, 0.719, 0.592,
		0.543, 0.021, 0.978,
		0.279, 0.317, 0.505,
		0.167, 0.620, 0.077,
		0.347, 0.857, 0.137,
		0.055, 0.953, 0.042,
		0.714, 0.505, 0.345,
		0.783, 0.290, 0.734,
		0.722, 0.645, 0.174,
		0.302, 0.455, 0.848,
		0.225, 0.587, 0.040,
		0.517, 0.713, 0.338,
		0.053, 0.959, 0.120,
		0.393, 0.621, 0.362,
		0.673, 0.211, 0.457,
		0.820, 0.883, 0.371,
		0.982, 0.099, 0.879,
	}

	var cbo uint32
	gl.GenBuffers(1, &cbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(cubeColors)*4, gl.Ptr(cubeColors), gl.STATIC_DRAW)

	var triangleVertices = []float32{
		-1.0, -1.0, 0.0,
		1.0, -1.0, 0.0,
		0.0, 1.0, 0.0,
	}

	var vboTriangle uint32
	gl.GenBuffers(1, &vboTriangle)
	gl.BindBuffer(gl.ARRAY_BUFFER, vboTriangle)
	gl.BufferData(gl.ARRAY_BUFFER, len(triangleVertices)*4, gl.Ptr(triangleVertices), gl.STATIC_DRAW)

	for !window.ShouldClose() && window.GetKey(glfw.KeyEscape) != glfw.Press {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render
		gl.UseProgram(program)

		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])
		gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		// Draw the cube
		gl.EnableVertexAttribArray(0)
		gl.BindBuffer(gl.ARRAY_BUFFER, vboCube)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3)
		gl.DisableVertexAttribArray(0)
		gl.DisableVertexAttribArray(1)

		// Draw the triangle
		gl.UniformMatrix4fv(modelUniform, 1, false, &model2[0])
		gl.EnableVertexAttribArray(0)
		gl.BindBuffer(gl.ARRAY_BUFFER, vboTriangle)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.EnableVertexAttribArray(1)
		gl.BindBuffer(gl.ARRAY_BUFFER, cbo)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		gl.DisableVertexAttribArray(0)
		gl.DisableVertexAttribArray(1)

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}

	gl.DeleteBuffers(1, &vboCube)
	gl.DeleteBuffers(1, &vboTriangle)
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteProgram(program)
}

var vertexShader = `
#version 330

uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;

layout(location = 0) in vec3 vert;
layout(location = 1) in vec3 color;

out vec3 fragmentColor;

void main() {
    gl_Position = projection * view * model * vec4(vert, 1);
		fragmentColor = color;
}
` + "\x00"

var fragmentShader = `
#version 330

in vec3 fragmentColor;

out vec3 outputColor;

void main() {
    outputColor = fragmentColor;
}
` + "\x00"

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

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

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
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
