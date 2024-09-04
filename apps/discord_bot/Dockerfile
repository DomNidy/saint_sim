# * It is important to know that the names of directories here aren't entirely arbitrary
# * We name them in accordance with the modules defined in our go.work file at the root of our directory
FROM golang:1.21 AS builder

# We will copy necessary go modules to build this image into this dir
WORKDIR /build

# copy our go.work files from root of our repository
COPY --from=root go.work go.work.sum ./

#* we can drop the ./apps/api module inside that go.work file as we only need discord_bot
# if we don't drop this module from the workspace file, we will run into an error as we 
# don't copy the api application into this container at all
RUN go work edit -dropuse ./apps/api

# copy the shared packages we need into here
# pkg is an additional_context defined in docker-compose.yml
WORKDIR /build/pkg
COPY --from=pkg . .

# copy the discord_bot apps into /apps/discord_bot
# we first copy the go.mod and download our modules so we can take advantage of docker cache
WORKDIR /build/apps/discord_bot
COPY go.mod .
RUN go mod download

# copy the rest of our apps into /apps/discord_bot
COPY . .

# Build the saint-sim binary
RUN go build -o discord_bot main.go

FROM alpine:2.7 AS runner

WORKDIR /app

# copy the binary from previous stage
COPY --from=builder /build/apps/discord_bot/discord_bot .
# todo: move to more secure way of managing these, perhaps use docker-compose config
# copy the .env from project root
COPY --from=root .env . 

ENTRYPOINT ["./discord_bot"]

