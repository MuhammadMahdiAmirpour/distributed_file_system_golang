<!-- Header -->
<div align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=12,14,25,27&height=180&section=header&text=Distributed%20File%20System&fontSize=36&fontAlignY=35&animation=twinkling&fontColor=FFFFFF"/>
</div>

# Distributed File System in Go

A distributed file system implementation written in Go, providing scalable and reliable file storage capabilities.

## 🛠️ Technology Stack

- **Go** - Primary programming language
- **Docker** - Containerization
- **gRPC** - For inter-service communication
- **Protocol Buffers** - Data serialization

## 🏗️ Project Structure

```
distributed_file_system_golang/
├── cmd/
│   └── main.go
├── docker/
│   └── Dockerfile
├── internal/
│   ├── controller/
│   ├── handler/
│   └── model/
├── pkg/
├── proto/
├── docker-compose.yml
└── README.md
```

## 🚀 Getting Started

### Prerequisites
- Go 1.19 or higher
- Docker and Docker Compose

### Running the Project

1. **Clone the repository**
   ```bash
   git clone https://github.com/MuhammadMahdiAmirpour/distributed_file_system_golang.git
   cd distributed_file_system_golang
   ```

2. **Build and run with Docker**
   ```bash
   docker-compose up --build
   ```

## 🌟 Features

- **Distributed Storage**: Store files across multiple nodes
- **Fault Tolerance**: Handle node failures gracefully
- **Scalability**: Easy to add new storage nodes
- **Data Replication**: Maintain multiple copies for reliability

## 🎓 Acknowledgments

This project was developed with the help of the following tutorial:
- [Distributed File System in Go](https://www.youtube.com/watch?v=bymQakvTY40)

## 👨‍💻 Author

**Muhammad Mahdi Amirpour**
- GitHub: [@MuhammadMahdiAmirpour](https://github.com/MuhammadMahdiAmirpour)

---

<div align="center">
  <sub>Built with ❤️ by Muhammad Mahdi Amirpour</sub>
</div>

<!-- Footer -->
<div align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=12,14,25,27&height=100&section=footer"/>
</div>
