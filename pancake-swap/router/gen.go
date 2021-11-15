package router

//go:generate abigen --sol IERC20.sol --out IERC20.sol.go --pkg router
//go:generate abigen --sol IPancakeMigrator.sol --out IPancakeMigrator.sol.go --pkg router
//go:generate abigen --sol IPancakeRouter02.sol --out IPancakeRouter02.sol.go --pkg router
//go:generate abigen --sol IWETH.sol --out IWETH.sol.go --pkg router
