package coremgt

func ExtractPlugin(pluginRecord []string) (index string, pluginData *ElementManifest) {
	pluginData = new(ElementManifest)
	pluginData.elementType = pluginRecord[0]
	pluginData.Name = pluginRecord[1]
	index = pluginRecord[1]
	pluginData.Version = pluginRecord[2]
	return
}

func ExtractGroovy(pluginRecord []string) (index string, pluginData *ElementManifest) {
	pluginData = new(ElementManifest)
	pluginData.elementType = pluginRecord[0]
	pluginData.Name = pluginRecord[1]
	pluginData.commitID = pluginRecord[2]
	index = pluginRecord[1]
	return
}

func ExtractFeature(pluginRecord []string) (index string, pluginData *ElementManifest) {
	pluginData = new(ElementManifest)
	pluginData.elementType = pluginRecord[0]
	pluginData.Name = pluginRecord[1]
	index = pluginRecord[1]
	return
}
