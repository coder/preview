// useWebSocket.ts
import { useEffect, useRef, useState } from "react";

export function useWebSocket<T>(url: string, reconnectDelay = 3000) {
  const [message, setMessage] = useState<T | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);

  useEffect(() => {
    const connectWebSocket = () => {
      try {
        const ws = new WebSocket(url);
        wsRef.current = ws;
        setConnectionStatus('connecting');

        ws.onopen = () => {
          console.log("Connected to WebSocket");
          setConnectionStatus('connected');
          ws.send(JSON.stringify({}));
        };

        ws.onmessage = (event) => {
          try {
            const data: T = JSON.parse(event.data);
            console.log("Received message:", data);
            setMessage(data);
          } catch (err) {
            console.error("Invalid JSON from server:", event.data);
          }
        };

        ws.onerror = (event) => {
          console.error("WebSocket error:", event);
        };

        ws.onclose = (event) => {
          console.log(`WebSocket closed with code ${event.code}. Reason: ${event.reason}`);
          setConnectionStatus('disconnected');
          
          // Attempt to reconnect after delay
          if (reconnectTimeoutRef.current) {
            window.clearTimeout(reconnectTimeoutRef.current);
          }
          reconnectTimeoutRef.current = window.setTimeout(connectWebSocket, reconnectDelay);
        };
      } catch (error) {
        console.error("Failed to create WebSocket connection:", error);
        setConnectionStatus('disconnected');
        
        // Attempt to reconnect after delay
        if (reconnectTimeoutRef.current) {
          window.clearTimeout(reconnectTimeoutRef.current);
        }
        reconnectTimeoutRef.current = window.setTimeout(connectWebSocket, reconnectDelay);
      }
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        window.clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [url, reconnectDelay]);

  const sendMessage = (data: unknown) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    } else {
      console.warn("Cannot send message: WebSocket is not connected");
    }
  };

  return { message, sendMessage, connectionStatus };
}