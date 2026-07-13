# Run Pasta2_3 plain with the Docker benchmark CLI

```powershell
git clone https://github.com/hosseinabdinf/HHE-ARENA.git
Set-Location HHE-ARENA

./frameworks/hheFramework/benchmark-docker.ps1
1

Get-Content results/hhe-framework/plain/raw.jsonl -TotalCount 3
Get-Content results/hhe-framework/plain/report/report.csv -TotalCount 20
Get-Content results/hhe-framework/plain/report/memory.csv -TotalCount 20
```

```powershell
./frameworks/hheFramework/benchmark-docker.ps1 -Backend plain -Iterations 10 -Warmups 2
```
