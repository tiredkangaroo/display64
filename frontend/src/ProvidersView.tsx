import { useState } from "react";
import type { Provider } from "./types";
import { apiURL } from "./api";
import { waitForWindowClose } from "./utils";

interface ProvidersViewProps {
    providers: Provider[];
    setProviders: React.Dispatch<React.SetStateAction<Provider[]>>;
}
export function ProvidersView({ providers, setProviders }: ProvidersViewProps) {
    if (providers.length === 0) {
        return <></>
    }
    const [selectedProvider, setSelectedProvider] = useState<Provider>((providers.find((p) => p.is_current))!);
    
    async function onProviderChange(newName: string) {
        const provider = providers.find((p) => p.name === newName);
        if (!provider) {
            console.error("Provider not found:", newName);
            return;
        }
        if (!provider.authorized) {
            const w = window.open(provider.authorization_url);
            // possibly add spinner while waiting for the popup to close on the left of "Select a provider:"
            await waitForWindowClose(w);
        }
        const resp = await fetch(apiURL + `/providers/start?name=${provider.name}`, {
            method: 'PUT',
        })
        if (!resp.ok) {
            const body = await resp.text();
            console.error("start provider:", body);
            return;
        }
        setProviders((prev) => prev.map((p) => ({
            ...p,
            is_current: p.name === newName
        })));
        setSelectedProvider(provider);
    }
    
    return (
        <div className="mt-2 w-full flex flex-row items-center justify-end">
            <div className="flex flex-row mr-6 items-center gap-2">
            Select a provider:
            <select className="border-1 border-black px-2 py-1" value={selectedProvider.name} onChange={(v) => onProviderChange(v.target.value) }>
            {providers.map((provider) => (
                <ProviderView key={provider.name} provider={provider}/>
            ))}
            </select>
            </div>
        </div>
    )
}



export function ProviderView({ provider }: { provider: Provider }) {
    return (
        <option key={provider.name} value={provider.name}>
            {provider.name} ({provider.is_current ? "Current" : provider.authorized ? "Ready" : "Requires Authorization" })
        </option>
    )
}