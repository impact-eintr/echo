package com.sanlei.websocketchatroom;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.web.socket.config.annotation.EnableWebSocket;
import org.springframework.web.socket.server.standard.ServerEndpointExporter;

@EnableWebSocket
@SpringBootApplication
public class WebsocketChatRoomApplication {

    public static void main(String[] args) {
        SpringApplication.run(WebsocketChatRoomApplication.class, args);
    }

    @Bean
    public ServerEndpointExporter serverEndpointExporter() {
        return new ServerEndpointExporter();
    }
}
