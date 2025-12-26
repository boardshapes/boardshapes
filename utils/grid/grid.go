package grid

func ForNonDiagonalAdjacents(x, y uint16, maxX, maxY int, function func(x, y uint16)) {
	if y > 0 {
		function(x, y-1)
	}
	if x > 0 {
		function(x-1, y)
	}
	if x < uint16(maxX)-1 {
		function(x+1, y)
	}
	if y < uint16(maxY)-1 {
		function(x, y+1)
	}
}

func ForAdjacents(x, y uint16, maxX, maxY int, function func(x, y uint16)) {
	if y > 0 {
		if x > 0 {
			function(x-1, y-1)
		}
		function(x, y-1)
		if x < uint16(maxX)-1 {
			function(x+1, y-1)
		}
	}
	if x > 0 {
		function(x-1, y)
	}
	if x < uint16(maxX)-1 {
		function(x+1, y)
	}
	if y < uint16(maxY)-1 {
		if x > 0 {
			function(x-1, y+1)
		}
		function(x, y+1)
		if x < uint16(maxX)-1 {
			function(x+1, y+1)
		}
	}
}
