import { useState, useEffect } from 'react';

interface User {
  groups: string[];
}

export function useUsers(serverAddress: string, testdata: string) {
  const [users, setUsers] = useState<Record<string, User>>({});
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [fetchError, setFetchError] = useState<string | null>(null);

  useEffect(() => {
    setIsLoading(true);
    setFetchError(null);
    
    if (testdata !== "") {
      fetch(`http://${serverAddress}/users/${testdata}`, {
        mode: 'cors',
        headers: {
          'Accept': 'application/json'
        }
      })
        .then(response => {
          if (!response.ok) {
            throw new Error(`Failed to fetch users: ${response.status} ${response.statusText}`);
          }
          return response.json();
        })
        .then(data => {
          setUsers(data);
          setIsLoading(false);
        })
        .catch(error => {
          console.error('Error fetching users:', error);
          setFetchError(error.message);
          setIsLoading(false);
        });
    }
  }, [serverAddress, testdata]);

  return { users, isLoading, fetchError };
} 