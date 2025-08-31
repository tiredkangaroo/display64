import { useEffect, useState } from "react";
import "./App.css";
import { apiURL } from "./api";
import type { Provider } from "./types";
import { ProvidersView } from "./ProvidersView";

function App() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [imageURL, setImageURL] = useState<string | null>(null);

  useEffect(() => {
    const ws = new WebSocket(apiURL + "/imageURL");
    ws.onmessage = (ev) => {
      setImageURL(ev.data);
    };

    fetch(apiURL + "/providers")
      .then((response) => response.json())
      .then((data) => {
        setProviders(data);
      })
      .catch((error) => {
        console.error("Error fetching providers:", error);
      });
  }, []);

  return (
    <div className="w-full h-full flex">
      <ProvidersView providers={providers} setProviders={setProviders} />
      <div className="w-full h-full flex items-center justify-center">
        {imageURL ? (
          <img src={imageURL} alt="Generated" width={640} height={640} />
        ) : (
          <p>No image available</p>
        )}
      </div>
    </div>
  );
}

export default App;
