import {useEffect, useState} from 'react';
import './App.css';
import {CreateInfrastructure, DestroyInfrastructure} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime/runtime";
import {Box, Paper, Stack, Switch, Typography} from "@mui/material";
import {styled} from "@mui/material/styles";

const OrchestrationSwitch = styled(Switch)(() => ({
    width: 128 * 3,
    height: 62 * 3,
    padding: 6 * 3,
    "& .MuiSwitch-switchBase": {
        padding: 6 * 3,
        transitionDuration: "300ms",
        "&.Mui-checked": {
            transform: "translateX(calc(70px * 3))",
            color: "#0f2c3a",
            "& + .MuiSwitch-track": {
                backgroundColor: "#9fe6ff",
                opacity: 1,
            },
        },
    },
    "& .MuiSwitch-thumb": {
        boxSizing: "border-box",
        width: 50 * 3,
        height: 50 * 3,
        backgroundColor: "#f7fbff",
        boxShadow: "0 8px 18px rgba(20, 60, 80, 0.25)",
        border: "1px solid rgba(15, 44, 58, 0.08)",
        position: "relative",
        "&::before": {
            content: "\"\"",
            position: "absolute",
            width: "58%",
            height: "58%",
            top: "21%",
            left: "21%",
            backgroundImage:
                "url(\"data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2024%2024'%20fill='none'%20stroke='%230f2c3a'%20stroke-width='2'%20stroke-linecap='round'%20stroke-linejoin='round'%3E%3Cpath%20d='M12%202v10'/%3E%3Cpath%20d='M7.5%204.5a8%208%200%201%200%209%200'/%3E%3C/svg%3E\")",
            backgroundRepeat: "no-repeat",
            backgroundPosition: "center",
            backgroundSize: "contain",
        },
    },
    "& .MuiSwitch-track": {
        borderRadius: 999,
        backgroundColor: "#d7effa",
        opacity: 1,
        boxShadow: "inset 0 0 0 1px rgba(15, 44, 58, 0.08)",
    },
}));

function App() {
    const [resultText, setResultText] = useState("");
    const [logLines, setLogLines] = useState<string[]>([]);
    const [isEnabled, setIsEnabled] = useState(false);
    const [isBusy, setIsBusy] = useState(false);

    useEffect(() => {
        const cancel = EventsOn("orchestrationLog", (message: string) => {
            setLogLines((prev) => [...prev, message]);
        });

        return () => cancel();
    }, []);

    async function toggleInfrastructure(nextEnabled: boolean) {
        if (isBusy) {
            return;
        }

        const previous = isEnabled;
        setIsEnabled(nextEnabled);
        setIsBusy(true);

        try {
            if (nextEnabled) {
                setResultText("Creating infrastructure, please wait...");
                const result = await CreateInfrastructure();
                setResultText(`Infrastructure created! Deployment URL: ${result}`);
            } else {
                setResultText("Destroying infrastructure, please wait...");
                await DestroyInfrastructure();
                setResultText("Infrastructure destroyed!");
            }
        } catch (error) {
            setIsEnabled(previous);
            setResultText("Operation failed. Check logs for details.");
        } finally {
            setIsBusy(false);
        }
    }

    return (
        <div id="App">
            <Paper className="orchestration-card" elevation={0}>
                <Stack spacing={2.5} alignItems="center">
                    <Typography variant="h4" className="headline">
                        DeVILStarter
                    </Typography>
                    <Typography variant="body2" className="subhead">
                        {isEnabled ? "Click to destroy infrastructure" : "Click to create infrastructure"}
                    </Typography>
                    <Box className="toggle-shell">
                        <OrchestrationSwitch
                            checked={isEnabled}
                            disabled={isBusy}
                            onChange={(event) => toggleInfrastructure(event.target.checked)}
                        />
                    </Box>
                    <Typography className="status-text" variant="body2">
                        {resultText}
                    </Typography>
                    <div id="orchestration-log" className="log-panel">
                        {logLines.length === 0 ? "Orchestration logs will appear here." : logLines.join("\n")}
                    </div>
                </Stack>
            </Paper>
        </div>
    )
}

export default App
