// git-recent-changes
// Copyright (C) 2022  Honza Pokorny <honza@pokorny.ca>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package recentchanges

import (
	"fmt"
	"testing"
)

const IND = "        "

func TestFormatBody(t *testing.T) {
	body := "This is a short line.\r\n\r\nThis is a second line."
	result := FormatBody(body, 8)
	expected := IND + "This is a short line.\n\n" + IND + "This is a second line."

	if result != expected {
		fmt.Println("---------------------------")
		fmt.Println(result)
		fmt.Println("---------------------------")
		fmt.Println(expected)
		t.Fail()
	}

}

func TestFormatBodyLongLine(t *testing.T) {
	body := "This is a long line that should definitely be wrapped so that it looks nicer in the output."
	result := FormatBody(body, 8)
	expected := IND + "This is a long line that should definitely be wrapped so that it looks\n" + IND + "nicer in the output."

	if result != expected {
		fmt.Println("---------------------------")
		fmt.Println(result)
		fmt.Println("---------------------------")
		fmt.Println(expected)
		t.Fail()
	}

}

func TestFormatBodyBlock(t *testing.T) {
	body := "This.\r\n\r\n```\r\nconsole.log('hi');\r\n```\r\n\r\nThat."
	result := FormatBody(body, 8)
	expected := IND + "This.\n\n" + IND + IND + "console.log('hi');\n\n" + IND + "That."

	if result != expected {
		fmt.Println("---------------------------")
		fmt.Println(result)
		fmt.Println("---------------------------")
		fmt.Println(expected)
		t.Fail()
	}

}
