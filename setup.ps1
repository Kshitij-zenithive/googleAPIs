# Create directories
New-Item -ItemType Directory -Path "cmd","internal","internal\config","internal\handler","internal\repository","internal\service","internal\domain","internal\web","internal\web\templates","internal\web\js" -Force

# Create files in each folder
New-Item -ItemType File -Path "cmd\main.go" -Force
New-Item -ItemType File -Path "internal\config\config.go" -Force
New-Item -ItemType File -Path "internal\handler\auth.go" -Force
New-Item -ItemType File -Path "internal\handler\event.go" -Force
New-Item -ItemType File -Path "internal\handler\handler.go" -Force
New-Item -ItemType File -Path "internal\handler\middleware.go" -Force
New-Item -ItemType File -Path "internal\repository\repository.go" -Force
New-Item -ItemType File -Path "internal\repository\user.go" -Force
New-Item -ItemType File -Path "internal\repository\meeting.go" -Force
New-Item -ItemType File -Path "internal\service\auth.go" -Force
New-Item -ItemType File -Path "internal\service\event.go" -Force
New-Item -ItemType File -Path "internal\service\service.go" -Force
New-Item -ItemType File -Path "internal\domain\models.go" -Force
New-Item -ItemType File -Path "internal\web\templates\dashboard.html" -Force
New-Item -ItemType File -Path "internal\web\templates\login.html" -Force
New-Item -ItemType File -Path "internal\web\js\dashboard.js" -Force
New-Item -ItemType File -Path "go.mod" -Force
New-Item -ItemType File -Path "go.sum" -Force
New-Item -ItemType File -Path ".env.example" -Force
New-Item -ItemType File -Path "wire.go" -Force
New-Item -ItemType File -Path "wire_gen.go" -Force
