<?php
namespace c;

use a;
use b;

const A = 1;
const B = "hello";
const C = 3.14;
const D = true;

enum Color : int
{
    case RED = 0;
    case GREEN = 1;
    case BLUE = 2;
}

enum LoginType : int
{
    case USERNAME = 1;
    case EMAIL = 2;
}

enum UserType : int
{
    case ADMIN = 1;
    case USER = 2;
}

class User
{
    public UserType $type;
    public int $id;
    public string $username;
    public string $password;
    public string $deviceId;
    public string $twoFactorToken;
    public array $roles;
    public array $metadata;
    public array $scores;

    public function __construct()
    {
        $this->type = UserType::ADMIN;
        $this->id = 0;
        $this->username = "";
        $this->password = "";
        $this->deviceId = "";
        $this->twoFactorToken = "";
        $this->roles = [];
        $this->metadata = [];
        $this->scores = [];
    }
}

class LoginRequest
{
    public LoginType $type;
    public string $username;
    public string $password;
    public string $deviceId;
    public string $twoFactorToken;

    public function __construct()
    {
        $this->type = LoginType::USERNAME;
        $this->username = "";
        $this->password = "";
        $this->deviceId = "";
        $this->twoFactorToken = "";
    }
}

class LoginResponse
{
    public bool $success;
    public string $errorMessage;
    public string $authenticationToken;
    public User $user;

    public function __construct()
    {
        $this->success = false;
        $this->errorMessage = "";
        $this->authenticationToken = "";
        $this->user = new User();
    }
}
