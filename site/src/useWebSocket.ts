// useWebSocket.ts
import { useEffect, useRef, useState, useCallback } from "react";

export function useWebSocket<T>(url: string, reconnectDelay = 3000) {
  const [message, setMessage] = useState<T | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const urlRef = useRef(url);
  
  // Update the URL ref when the URL changes
  useEffect(() => {
    urlRef.current = url;
  }, [url]);

  // Memoize the connect function to avoid recreating it on every render
  const connectWebSocket = useCallback(() => {
    try {
      // Always use the latest URL from the ref
      const ws = new WebSocket(urlRef.current);
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
  }, [reconnectDelay]); // Only depends on reconnectDelay, not url

  // Set up the WebSocket connection
  useEffect(() => {
    // Close any existing connection when the URL changes
    if (wsRef.current) {
      wsRef.current.close();
    }
    
    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        window.clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [url, connectWebSocket]); // Reconnect when URL changes

  const sendMessage = (data: unknown) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));

    } else {
      console.warn("Cannot send message: WebSocket is not connected");
    }
  };

  return { message, sendMessage, connectionStatus };
}