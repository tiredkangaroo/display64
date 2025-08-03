import { useEffect, useState } from 'react'
import './App.css'
import { apiURL } from './api';
import type { Provider } from './types';
import { ProvidersView } from './ProvidersView';


function App() {
  const [providers, setProviders] = useState<Provider[]>([  ]);
  
  useEffect(() => {
    fetch(apiURL + "/providers")
      .then(response => response.json())
      .then(data => {
        setProviders(data);
      })
      .catch(error => {
        console.error('Error fetching providers:', error);
      });
  }, [])

  return (
    <div className="w-full h-full">
      <ProvidersView providers={providers} setProviders={setProviders} />
    </div>
  )
}

export default App
