package contract

//go:generate abigen --sol IERC20.sol --out IERC20.sol.go --pkg contract
//go:generate abigen --sol IPancakeERC20.sol --out IPancakeERC20.sol.go --pkg contract
//go:generate abigen --sol IPancakeFactory.sol --out IPancakeFactory.sol.go --pkg contract
//go:generate abigen --sol IPancakePair.sol --out IPancakePair.sol.go --pkg contract
