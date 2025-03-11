import { useState, useEffect } from 'react';

export function useDirectories(serverAddress: string, initialTestdata: string) {
  const [directories, setDirectories] = useState<string[]>([]);
  const [testdata, setTestdata] = useState<string>(initialTestdata);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [fetchError, setFetchError] = useState<string | null>(null);

  useEffect(() => {
    setIsLoading(true);
    setFetchError(null);
    
    // Use mode: 'cors' explicitly and add credentials if needed
    fetch(`http://${serverAddress}/directories`, {
      mode: 'cors',
      headers: {
        'Accept': 'application/json'
      }
    })
      .then(response => {
        if (!response.ok) {
          throw new Error(`Failed to fetch directories: ${response.status} ${response.statusText}`);
        }
        return response.json();
      })
      .then(data => {
        setDirectories(data);
        // If testdata is not in the list of directories, set it to the first directory
        if (data.length > 0 && !data.includes(testdata)) {
          setTestdata(data[0]);
        }
        setIsLoading(false);
      })
      .catch(error => {
        console.error('Error fetching directories:', error);
        setFetchError(error.message);
        // Fallback to some default directories if fetch fails
        setDirectories(["conditional"]);
        setIsLoading(false);
      });
  }, [serverAddress, testdata]);

  return { directories, testdata, setTestdata, isLoading, fetchError };
} 