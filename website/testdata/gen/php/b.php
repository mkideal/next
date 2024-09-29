<?php
namespace b;

use a;


enum TestEnum : int
{
    case A = 1;
    case B = 5;
    case C = 5;
    case D = 10;
    case E = 20;
    case F = 1;
    case G = 2;
}

class TestStruct
{
    public \a\Point2D $point;

    public function __construct()
    {
        $this->point = new \a\Point2D();
    }
}
